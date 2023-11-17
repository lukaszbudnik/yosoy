package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const PING_DEFAULT_TIMEOUT = 10
const PING_DEFAULT_NETWORK = "tcp"

type response struct {
	Host         string              `json:"host"`
	Proto        string              `json:"proto"`
	Method       string              `json:"method"`
	Scheme       string              `json:"scheme"`
	RequestURI   string              `json:"requestUri"`
	URL          string              `json:"url"`
	RemoteAddr   string              `json:"remoteAddr"`
	Counter      int                 `json:"counter"`
	Headers      map[string][]string `json:"headers"`
	Hostname     string              `json:"hostname"`
	EnvVariables []string            `json:"envVariables,omitempty"`
	Files        map[string]string   `json:"files,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type successResponse struct {
	Message string `json:"message"`
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
	response.Proto = req.Proto
	response.Method = req.Method
	response.Scheme = req.URL.Scheme
	response.Headers = req.Header
	response.URL = req.URL.String()

	response.Hostname = hostname

	if strings.ToLower(showEnvs) == "true" || strings.ToLower(showEnvs) == "yes" || strings.ToLower(showEnvs) == "on" || showEnvs == "1" {
		response.EnvVariables = os.Environ()
	}
	if len(showFiles) > 0 {
		response.Files = make(map[string]string)
		files := strings.Split(showFiles, ",")
		for _, file := range files {
			bytes, err := os.ReadFile(file)
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

func ping(w http.ResponseWriter, req *http.Request) {
	// get h, p, n, t parameters from query string
	hostname := req.URL.Query().Get("h")
	port := req.URL.Query().Get("p")
	network := req.URL.Query().Get("n")
	timeoutString := req.URL.Query().Get("t")
	var timeout int64 = PING_DEFAULT_TIMEOUT

	// return HTTP BadRequest when hostname is empty
	if hostname == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(&errorResponse{"hostname is empty"})
		return
	}
	// return HTTP BadRequest when port is empty
	if port == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(&errorResponse{"port is empty"})
		return
	}
	// check if timeoutString is a valid int64, return HTTP BadRequest when invalid, otherwise set timeout value
	if timeoutString != "" {
		timeoutInt, err := strconv.ParseInt(timeoutString, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Add("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(&errorResponse{"timeout is invalid"})
			return
		}
		timeout = timeoutInt
	}

	// if network is empty set default to tcp
	if network == "" {
		network = PING_DEFAULT_NETWORK
	}

	// ping the hostname and port
	err := pingHost(hostname, port, network, timeout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(&errorResponse{err.Error()})
		return
	}
	// return HTTP OK
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(&successResponse{"ping succeeded"})
}

func pingHost(hostname, port, network string, timeout int64) error {
	// create timeoutDuration variable of Duration type using timeout as the value in seconds
	timeoutDuration := time.Duration(timeout) * time.Second

	// open a socket to hostname and port
	conn, err := net.DialTimeout(network, hostname+":"+port, timeoutDuration)
	if err != nil {
		return err
	}
	// close the socket
	conn.Close()
	return nil
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
	r.HandleFunc("/_/yosoy/ping", ping).Methods(http.MethodGet)
	r.PathPrefix("/").HandlerFunc(preflight).Methods(http.MethodOptions)
	r.PathPrefix("/").HandlerFunc(handler).Methods(http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodConnect, http.MethodHead, http.MethodTrace)

	loggingRouter := handlers.CombinedLoggingHandler(os.Stdout, r)
	proxyRouter := handlers.ProxyHeaders(loggingRouter)
	recoveryRouter := handlers.RecoveryHandler()(proxyRouter)

	http.ListenAndServe(":80", recoveryRouter)
}
