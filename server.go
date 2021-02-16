package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type response struct {
	Host         string              `json:"host"`
	RequestURI   string              `json:"requestUri"`
	RemoteAddr   string              `json:"remoteAddr"`
	Counter      int                 `json:"counter"`
	Headers      map[string][]string `json:"headers"`
	Hostname     string              `json:"hostname"`
	EnvVariables []string            `json:"envVariables,omitempty"`
	Files        map[string]string   `json:"files,omitempty"`
}

var counter = 0
var hostname = os.Getenv("HOSTNAME")

func preflight(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Expose-Headers", "*")
	w.Header().Set("Access-Control-Max-Age", "600")
	w.WriteHeader(http.StatusOK)
}

func handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	showEnvs := os.Getenv("YOSOY_SHOW_ENVS")
	showFiles := os.Getenv("YOSOY_SHOW_FILES")

	response := &response{}

	counter++
	response.Counter = counter

	remoteAddr := remoteAddrWithoutPort(req)
	response.RemoteAddr = remoteAddr

	response.RequestURI = req.RequestURI
	response.Host = req.Host
	response.Headers = req.Header

	response.Hostname = hostname

	if strings.ToLower(showEnvs) == "true" || strings.ToLower(showEnvs) == "yes" || strings.ToLower(showEnvs) == "on" || showEnvs == "1" {
		response.EnvVariables = os.Environ()
	}
	if len(showFiles) > 0 {
		response.Files = make(map[string]string)
		files := strings.Split(showFiles, ",")
		for _, file := range files {
			bytes, err := ioutil.ReadFile(file)
			if err != nil {
				log.Printf("Could not read file %v: %v\n", file, err)
				continue
			}
			contents := string(bytes)
			response.Files[file] = contents
		}
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func remoteAddrWithoutPort(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if index := strings.LastIndex(remoteAddr, ":"); index > 0 {
		remoteAddr = remoteAddr[0:index]
	}
	return remoteAddr
}

func main() {
	log.Printf("yosoy is up %v\n", hostname)

	r := mux.NewRouter()

	r.Handle("/favicon.ico", r.NotFoundHandler)
	r.PathPrefix("/").HandlerFunc(preflight).Methods(http.MethodOptions)
	r.PathPrefix("/").HandlerFunc(handler).Methods(http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodConnect, http.MethodHead, http.MethodTrace)

	loggingRouter := handlers.CombinedLoggingHandler(os.Stdout, r)
	proxyRouter := handlers.ProxyHeaders(loggingRouter)
	recoveryRouter := handlers.RecoveryHandler()(proxyRouter)

	http.ListenAndServe(":80", recoveryRouter)
}
