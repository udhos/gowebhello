package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	helloVersion = "0.1"
)

var knownPaths []string
var boottime time.Time
var banner string

func main() {

	boottime = time.Now()

	tls := true

	log.Print("version: ", helloVersion)
	log.Print("runtime: ", runtime.Version())

	currDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Getwd: %s", err)
	}

	var addr, httpsAddr, key, cert string

	flag.StringVar(&key, "key", "key.pem", "TLS key file")
	flag.StringVar(&cert, "cert", "cert.pem", "TLS cert file")
	flag.StringVar(&addr, "addr", ":8080", "HTTP listen address")
	flag.StringVar(&httpsAddr, "httpsAddr", ":8443", "HTTPS listen address")
	flag.StringVar(&banner, "banner", "deploy #4", "banner will be displayed")
	flag.Parse()

	log.Print("banner: ", banner)

	if !fileExists(key) {
		log.Printf("TLS key file not found: %s - disabling TLS", key)
		tls = false
	}

	if !fileExists(cert) {
		log.Printf("TLS cert file not found: %s - disabling TLS", cert)
		tls = false
	}

	http.HandleFunc("/", rootHandler) // default handler

	registerStatic("/www/", currDir)

	log.Printf("using TCP ports HTTP=%s HTTPS=%s TLS=%v", addr, httpsAddr, tls)

	if tls {
		serveHTTPS(addr, httpsAddr, cert, key)
		return
	}

	log.Printf("serving HTTP on TCP %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("ListenAndServe: %s: %v", addr, err)
	}
}

func serveHTTPS(addr, httpsAddr, cert, key string) {
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
			if err := http.ListenAndServe(addr, http.HandlerFunc(redirectTLS)); err != nil {
				log.Fatalf("redirect: ListenAndServe: %s: %v", addr, err)
			}
		}()
	}

	log.Printf("serving HTTPS on TCP %s", httpsAddr)
	if err := http.ListenAndServeTLS(httpsAddr, cert, key, nil); err != nil {
		log.Fatalf("ListenAndServeTLS: %s: %v", httpsAddr, err)
	}
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
	log.Printf("registering static directory %s as www path %s", dir, path)
}

func (handler staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("staticHandler.ServeHTTP url=%s from=%s", r.URL.Path, r.RemoteAddr)
	handler.innerHandler.ServeHTTP(w, r)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("rootHandler: url=%s from=%s", r.URL.Path, r.RemoteAddr)
	log.Print(msg)

	var paths string
	for _, p := range knownPaths {
		paths += fmt.Sprintf("<a href=\"%s\">%s</a> <br>", p, p)
	}

	var errMsg string
	if r.URL.Path != "/" {
		errMsg = fmt.Sprintf("<h2>Path not found!</h2>Path not found: [%s]", r.URL.Path)
	}

	rootStr :=
		`<!DOCTYPE html>

<html>
  <head>
    <title>gowebhello root page</title>
  </head>
  <body>
    <h1>gowebhello root page</h1>
    <p>
    <a href="https://github.com/udhos/gowebhello">gowebhello</a> is a simple golang replacement for 'python -m SimpleHTTPServer'.
    </p>
    <h2>Welcome!</h2>
	gowebhello version: %s<br>
	Golang version: %s<br>
	Application banner: %s<br>
	Application arguments: %v<br>
	Application dir: %s<br>
	Server hostname: %s<br>
	Your address: %s<br>
	Current time: %s<br>
	Uptime: %s<br>
    %s
    <h2>All known paths:</h2>
    %s
  </body>
</html>
`

	cwd, errCwd := os.Getwd()
	if errCwd != nil {
		cwd = cwd + " (error: " + errCwd.Error() + ")"
	}

	host, errHost := os.Hostname()
	if errHost != nil {
		host = host + " (error: " + errHost.Error() + ")"
	}

	now := time.Now()

	rootPage := fmt.Sprintf(rootStr, helloVersion, runtime.Version(), banner, os.Args, cwd, host, r.RemoteAddr, now, time.Since(boottime), errMsg, paths)

	io.WriteString(w, rootPage)
}
