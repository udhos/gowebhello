package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

var knownPaths []string
var boottime time.Time

func main() {

	boottime = time.Now()

	tls := true

	currDir, err := os.Getwd()
	if err != nil {
		log.Panicf("Getwd: %s", err)
	}

	var addr, key, cert string

	flag.StringVar(&key, "key", "key.pem", "TLS key file")
	flag.StringVar(&cert, "cert", "cert.pem", "TLS cert file")
	flag.StringVar(&addr, "addr", ":8080", "listen address")
	flag.Parse()

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

	log.Printf("serving on port TCP %s TLS=%v", addr, tls)

	if tls {
		if err := http.ListenAndServeTLS(addr, cert, key, nil); err != nil {
			log.Panicf("ListenAndServeTLS: %s: %s", addr, err)
		}
		return
	}

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Panicf("ListenAndServe: %s: %s", addr, err)
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
	Golang version: %s<br>
	Application version: 3<br>
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

	rootPage := fmt.Sprintf(rootStr, runtime.Version(), os.Args, cwd, host, r.RemoteAddr, now, time.Since(boottime), errMsg, paths)

	io.WriteString(w, rootPage)
}
