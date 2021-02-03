package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var counter = 0
var hostname = os.Getenv("HOSTNAME")
var showEnvs = os.Getenv("YOSOY_SHOW_ENVS")
var showFiles = os.Getenv("YOSOY_SHOW_FILES")

func handler(w http.ResponseWriter, req *http.Request) {
	remoteAddr := req.RemoteAddr
	if index := strings.LastIndex(remoteAddr, ":"); index > 0 {
		remoteAddr = remoteAddr[0:index]
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "text/plain")
	fmt.Fprintf(w, "Request URI: %v\n", req.RequestURI)
	fmt.Fprintf(w, "Hostname: %v\n", hostname)
	fmt.Fprintf(w, "Remote IP: %v\n", remoteAddr)
	counter++
	fmt.Fprintf(w, "Called: %v\n", counter)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "HTTP headers:\n")
	headers := make([]string, 0, len(req.Header))
	for k := range req.Header {
		headers = append(headers, k)
	}
	sort.Strings(headers)
	for _, header := range headers {
		headers := req.Header[header]
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", header, h)
		}
	}
	if strings.ToLower(showEnvs) == "true" || strings.ToLower(showEnvs) == "yes" || strings.ToLower(showEnvs) == "on" || showEnvs == "1" {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "Env variables:\n")
		for _, e := range os.Environ() {
			fmt.Fprintln(w, e)
		}
	}
	if len(showFiles) > 0 {
		files := strings.Split(showFiles, ",")
		for _, file := range files {
			bytes, err := ioutil.ReadFile(file)
			if err != nil {
				log.Printf("Could not read file %v: %v\n", file, err)
				continue
			}
			fmt.Fprintln(w)
			fmt.Fprintf(w, "File %v:\n", file)
			contents := string(bytes)
			fmt.Fprintln(w, contents)
		}
	}
}

func main() {
	log.Printf("yosoy is up %v\n", hostname)

	r := mux.NewRouter()

	r.Handle("/favicon.ico", r.NotFoundHandler)
	r.PathPrefix("/").HandlerFunc(handler)

	loggingRouter := handlers.CombinedLoggingHandler(os.Stdout, r)
	proxyRouter := handlers.ProxyHeaders(loggingRouter)
	recoveryRouter := handlers.RecoveryHandler()(proxyRouter)

	http.ListenAndServe(":80", recoveryRouter)
}
