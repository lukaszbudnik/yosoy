package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

var counter = 0
var hostname = os.Getenv("HOSTNAME")
var showEnvs = os.Getenv("YOSOY_SHOW_ENVS")
var showFiles = os.Getenv("YOSOY_SHOW_FILES")

func handler(w http.ResponseWriter, req *http.Request) {
	if req.RequestURI == "/favicon.ico" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	remoteAddr := req.RemoteAddr
	if index := strings.LastIndex(remoteAddr, ":"); index > 0 {
		remoteAddr = remoteAddr[0:index]
	}
	fmt.Printf("[%v] - %v - %v - \"%v %v\"\n", hostname, time.Now().Format(time.RFC3339), remoteAddr, req.Method, req.RequestURI)
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
				fmt.Printf("[%v] - %v - could not read file %v: %v\n", hostname, time.Now().Format(time.RFC3339), file, err)
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
	fmt.Printf("[%v] - %v - yosoy is up!\n", hostname, time.Now().Format(time.RFC3339))
	http.HandleFunc("/", handler)
	http.ListenAndServe(":80", nil)
}
