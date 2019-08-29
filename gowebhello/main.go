package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"time"
	"unicode"
)

const (
	helloVersion = "0.8"
)

var knownPaths []string
var boottime time.Time
var banner string
var requests int64
var quota int64
var quotaDuration time.Duration
var exitOnQuota bool
var quotaStatus int
var usr *user.User
var burnCPU bool
var listenAddr, listenAddrHTTPS string

func inc() int64 {
	return atomic.AddInt64(&requests, 1)
}

func get() int64 {
	return atomic.LoadInt64(&requests)
}

func getVersion() string {
	return fmt.Sprintf("version=%s runtime=%s pid=%d host=%s GOMAXPROCS=%d OS=%s ARCH=%s", helloVersion, runtime.Version(), os.Getpid(), getHostname(), runtime.GOMAXPROCS(0), runtime.GOOS, runtime.GOARCH)
}

func getHostname() string {
	host, errHost := os.Hostname()
	if errHost != nil {
		host = host + " (host error: " + errHost.Error() + ")"
	}
	return host
}

func dumpVersion(path string) {
	log.Printf("dumpVersion: writing to touch=%s", path)
	v := getVersion()
	err := ioutil.WriteFile(path, []byte(v), 0640)
	if err != nil {
		log.Printf("dumpVersion: touch=%s: %v", path, err)
	}
}

func main() {

	boottime = time.Now()

	tls := true

	log.Print(getVersion())

	log.Printf("GWH_BANNER='%s' PORT='%s'", os.Getenv("GWH_BANNER"), os.Getenv("PORT"))

	defaultBanner := "banner default"
	if envBanner := os.Getenv("GWH_BANNER"); envBanner != "" {
		defaultBanner = envBanner
	}

	defaultAddr := ":8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		defaultAddr = ":" + envPort
	}

	var errUser error
	usr, errUser = user.Current()
	if errUser != nil {
		log.Printf("current user error: %v", errUser)
	}
	if usr != nil {
		log.Printf("user: %s (uid: %s)", usr.Username, usr.Uid)
	}

	currDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Getwd: %s", err)
	}

	var addr, httpsAddr, key, cert string
	var disableKeepalive bool
	var touch string
	var showVersion bool
	var quotaTime string

	flag.StringVar(&key, "key", "key.pem", "TLS key file")
	flag.StringVar(&cert, "cert", "cert.pem", "TLS cert file")
	flag.StringVar(&addr, "addr", defaultAddr, "HTTP listen address (env var PORT sets default port)")
	flag.StringVar(&httpsAddr, "httpsAddr", ":8443", "HTTPS listen address")
	flag.StringVar(&banner, "banner", defaultBanner, "banner will be displayed (env var GWH_BANNER sets default banner)")
	flag.StringVar(&touch, "touch", "", "write version info into this file, if specified")
	flag.BoolVar(&showVersion, "version", false, "does not actually run")
	flag.BoolVar(&disableKeepalive, "disableKeepalive", false, "disable keepalive")
	flag.Int64Var(&quota, "quota", 0, "if defined, service is terminated after serving that amount of requests")
	flag.StringVar(&quotaTime, "quotaTime", "", "if defined, service is terminated after serving that amount of time")
	flag.BoolVar(&exitOnQuota, "exitOnQuota", false, "exit if quota reached")
	flag.IntVar(&quotaStatus, "quotaStatus", 500, "http status code for quota reached")
	flag.BoolVar(&burnCPU, "burnCpu", false, "enable /burncpu path")
	flag.Parse()

	if touch != "" {
		dumpVersion(touch)
	}

	keepalive := !disableKeepalive

	if quotaTime != "" {
		// append "s" to all-numeric duration string
		last := len(quotaTime) - 1
		if unicode.IsDigit(rune(quotaTime[last])) {
			quotaTime += "s"
		}

		// parse quotaTime
		var errDur error
		quotaDuration, errDur = time.ParseDuration(quotaTime)
		if errDur != nil {
			log.Printf("quotaTime: %v", errDur)
		}
	}

	log.Print("banner: ", banner)
	log.Print("keepalive: ", keepalive)
	log.Printf("burnCpu=%v", burnCPU)
	log.Printf("quotaTime=%v", quotaDuration)
	log.Printf("requests quota=%d (0=unlimited) exitOnQuota=%v quotaStatus=%d", quota, exitOnQuota, quotaStatus)

	if !fileExists(key) {
		log.Printf("TLS key file not found: %s - disabling TLS", key)
		tls = false
	}

	if !fileExists(cert) {
		log.Printf("TLS cert file not found: %s - disabling TLS", cert)
		tls = false
	}

	if showVersion {
		return
	}

	// default handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { rootHandler(w, r, keepalive) })

	if burnCPU {
		log.Print("burnCpu: enabling path /burncpu")
		http.HandleFunc("/burncpu", func(w http.ResponseWriter, r *http.Request) { burncpuHandler(w, r, keepalive) })
		http.HandleFunc("/burncpu/", func(w http.ResponseWriter, r *http.Request) { burncpuHandler(w, r, keepalive) })
	}

	registerStatic("/www/", currDir)

	// save for html
	listenAddr = addr
	listenAddrHTTPS = httpsAddr

	log.Printf("using TCP ports HTTP=%s HTTPS=%s TLS=%v", addr, httpsAddr, tls)

	if tls {
		serveHTTPS(addr, httpsAddr, cert, key, keepalive)
		return
	}

	log.Printf("serving HTTP on TCP %s", addr)
	if err := listenAndServe(addr, nil, keepalive); err != nil {
		log.Fatalf("listenAndServe: %s: %v", addr, err)
	}
}

func serveHTTPS(addr, httpsAddr, cert, key string, keepalive bool) {
	httpPort := "80"
	h := strings.Split(addr, ":")
	if len(h) > 1 {
		httpPort = h[1]
	}

	httpsPort := "443"
	hs := strings.Split(httpsAddr, ":")
	if len(hs) > 1 {
		httpsPort = hs[1]
	}

	if httpPort != httpsPort {
		log.Printf("installing redirect from HTTP=%s to HTTPS=%s", addr, httpsPort)

		redirectTLS := func(w http.ResponseWriter, r *http.Request) {
			host := strings.Split(r.Host, ":")[0]
			//code := http.StatusMovedPermanently // 301 permanent
			code := http.StatusFound // 302 temporary
			http.Redirect(w, r, "https://"+host+":"+httpsPort+r.RequestURI, code)
		}

		// http-to-https redirect server
		go func() {
			if err := listenAndServe(addr, http.HandlerFunc(redirectTLS), keepalive); err != nil {
				log.Fatalf("redirect: listenAndServe: %s: %v", addr, err)
			}
		}()
	}

	log.Printf("serving HTTPS on TCP %s", httpsAddr)
	if err := listenAndServeTLS(httpsAddr, cert, key, nil, keepalive); err != nil {
		log.Fatalf("listenAndServeTLS: %s: %v", httpsAddr, err)
	}
}

func listenAndServe(addr string, handler http.Handler, keepalive bool) error {
	server := &http.Server{Addr: addr, Handler: handler}
	server.SetKeepAlivesEnabled(keepalive)
	return server.ListenAndServe()
}

func listenAndServeTLS(addr, certFile, keyFile string, handler http.Handler, keepalive bool) error {
	server := &http.Server{Addr: addr, Handler: handler}
	server.SetKeepAlivesEnabled(keepalive)
	return server.ListenAndServeTLS(certFile, keyFile)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

type staticHandler struct {
	innerHandler http.Handler
}

func registerStatic(path, dir string) {
	http.Handle(path, staticHandler{http.StripPrefix(path, http.FileServer(http.Dir(dir)))})
	knownPaths = append(knownPaths, path)
	log.Printf("mapping www path %s to directory %s", path, dir)
}

func (handler staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	count := inc()
	log.Printf("staticHandler.ServeHTTP req=%d url=%s from=%s", count, r.URL.Path, r.RemoteAddr)
	if reached := checkQuota("staticHandler.ServeHTTP", count); reached {
		quotaError(w)
		return
	}
	handler.innerHandler.ServeHTTP(w, r)
}

func quotaError(w http.ResponseWriter) {
	http.Error(w, "Quota forced malfunction", quotaStatus)
}

func checkQuota(label string, count int64) bool {
	if quota > 0 && count > quota {
		if exitOnQuota {
			log.Printf("%s req=%d quota=%d exitOnQuota=%v quotaStatus=%d reached EXITING", label, count, quota, exitOnQuota, quotaStatus)
			os.Exit(0)
		}
		log.Printf("%s req=%d quota=%d exitOnQuota=%v quotaStatus=%d reached", label, count, quota, exitOnQuota, quotaStatus)
		return true
	}
	if quotaDuration > 0 {
		uptime := time.Since(boottime)
		timeLeft := quotaDuration - uptime
		if timeLeft <= 0 {
			if exitOnQuota {
				log.Printf("%s req=%d exitOnQuota=%v uptime=%v quotaTime=%v timeLeft=%v reached EXITING", label, count, exitOnQuota, uptime, quotaDuration, timeLeft)
				os.Exit(0)
			}
			log.Printf("%s req=%d exitOnQuota=%v uptime=%v quotaTime=%v timeLeft=%v reached", label, count, exitOnQuota, uptime, quotaDuration, timeLeft)
			return true
		}
	}
	return false
}

func burncpuHandler(w http.ResponseWriter, r *http.Request, keepalive bool) {
	count := inc()
	log.Printf("burncpuHandler: req=%d from=%s", count, r.RemoteAddr)

	if reached := checkQuota("burncpuHandler", count); reached {
		quotaError(w)
		return
	}

	burn()

	io.WriteString(w, "done\n")
}

func burn() {
	for i := 0; i < 2000000000; i++ {
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request, keepalive bool) {
	count := inc()
	log.Printf("rootHandler: req=%d from=%s", count, r.RemoteAddr)
	log.Printf("rootHandler: req=%d method=%s host=%s path=%s", count, r.Method, r.Host, r.URL.Path)
	log.Printf("rootHandler: req=%d query=%s", count, r.URL.RawQuery)

	if reached := checkQuota("rootHandler", count); reached {
		quotaError(w)
		return
	}

	var paths string
	for _, p := range knownPaths {
		paths += fmt.Sprintf("<a href=\"%s\">%s</a> <br>", p, p)
	}

	var errMsg string
	if r.URL.Path != "/" {
		format := `
<h2>Path not found!</h2>
Method: %s<br>
Host: %s<br>
Path not found: [%s]<br>
Query: [%s]<br>
`
		errMsg = fmt.Sprintf(format, r.Method, r.Host, r.URL.Path, r.URL.RawQuery)
	}

	header :=
		`<!DOCTYPE html>

<html>
  <head>
    <title>gowebhello root page</title>
  </head>
  <body>
    <h1>gowebhello root page</h1>
    <p>
    <a href="https://github.com/udhos/gowebhello">https://github.com/udhos/gowebhello</a> is a simple golang replacement for 'python -m SimpleHTTPServer'.
    </p>
`
	bodyTempl :=
		`<h2>Welcome!</h2>
	gowebhello version %s runtime %s os=%s arch=%s<br>
        Keepalive: %v<br>
	Application banner: %s<br>
	Application arguments: %v<br>
	Application dir: %s<br>
	Process: %d<br>
	User: %s (uid: %s)<br>
	Server hostname: %s<br>
	Listen: HTTP=%s HTTPS=%s<br>
	Your address: %s<br>
	Request method=%s host=%s path=[%s] query=[%s]<br>
	Current time: %s<br>
	Uptime: %s<br>
	Requests: %d (Quota=%d ExitOnQuota=%v QuotaStaus=%d QuotaTime=%v TimeLeft=%v)<br>
	BurnCPU: %v<br>
    %s
    <h2>All known paths:</h2>
    %s
`

	footer :=
		`</body>
</html>
`

	cwd, errCwd := os.Getwd()
	if errCwd != nil {
		cwd = cwd + " (error: " + errCwd.Error() + ")"
	}

	host := getHostname()

	now := time.Now()

	username := "?"
	uid := "?"

	if usr != nil {
		username = usr.Username
		uid = usr.Uid
	}

	uptime := time.Since(boottime)
	timeLeft := quotaDuration - uptime
	if timeLeft < 0 {
		timeLeft = 0
	}

	body := fmt.Sprintf(bodyTempl, helloVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH, keepalive, banner, os.Args, cwd, os.Getpid(), username, uid, host, listenAddr, listenAddrHTTPS, r.RemoteAddr, r.Method, r.Host, r.URL.Path, r.URL.RawQuery, now, uptime, get(), quota, exitOnQuota, quotaStatus, quotaDuration, timeLeft, burnCPU, errMsg, paths)

	if !keepalive {
		w.Header().Set("Connection", "close")
	}

	io.WriteString(w, header)
	io.WriteString(w, body)
	showHeaders(w, r)
	showReqBody(w, r)
	io.WriteString(w, footer)
}

func showReqBody(w http.ResponseWriter, r *http.Request) {

	io.WriteString(w, "<h2>Request Body</h2>\n")

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		m := fmt.Sprintf("request body: %v", err)
		log.Print(m)
		io.WriteString(w, m)
		return
	}

	bodySize := len(buf)

	log.Printf("body size: %d", bodySize)

	io.WriteString(w, fmt.Sprintf("<p>request body size: %d</p>\n", bodySize))

	io.WriteString(w, "<pre>\n")
	io.WriteString(w, string(buf))
	io.WriteString(w, "\n")
	io.WriteString(w, "</pre>\n")
}

func showHeaders(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "<h2>Headers</h2>\n")

	var headers []string
	for k, v := range r.Header {
		for _, vv := range v {
			headers = append(headers, k+": "+vv+"<br>\n")
		}
	}

	sort.Strings(headers)

	for _, h := range headers {
		io.WriteString(w, h)
	}
}
