package main

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	os.Setenv("YOSOY_SHOW_ENVS", "true")
	os.Setenv("YOSOY_SHOW_FILES", ".gitignore,/file/does/not/exist")

	req, err := http.NewRequest("GET", "https://example.org/sample/path?one=jeden&two=dwa", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "*/*")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	result := rr.Result()
	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)

	var response response
	json.Unmarshal(buf.Bytes(), &response)

	// test response
	assert.Equal(t, 1, response.Counter)
	assert.Equal(t, "example.org", response.Host)
	assert.Equal(t, "GET", response.Method)
	assert.Equal(t, "https", response.Scheme)
	assert.Equal(t, "HTTP/1.1", response.Proto)
	assert.Equal(t, "https://example.org/sample/path?one=jeden&two=dwa", response.URL)
	assert.NotEmpty(t, response.EnvVariables)
	assert.NotEmpty(t, response.Files[".gitignore"])

	// test cors
	assert.Contains(t, result.Header["Access-Control-Allow-Origin"], "*")
}

// write test for request /_/yosoy/ping without any query parameters, the request should return bad request 400 error and return JSON error about missing hostname parameter
func TestHandlerPingNoParameters(t *testing.T) {
	req, err := http.NewRequest("GET", "https://example.org/_/yosoy/ping", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "*/*")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ping)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	result := rr.Result()
	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)

	var response errorResponse
	json.Unmarshal(buf.Bytes(), &response)

	// test response
	assert.Equal(t, "hostname is empty", response.Error)
}

// write test for request /_/yosoy/ping with h parameter, the request should return bad request 400 error and return JSON error about port is empty
func TestHandlerPingWithHostname(t *testing.T) {
	req, err := http.NewRequest("GET", "https://example.org/_/yosoy/ping?h=example.org", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "*/*")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ping)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	result := rr.Result()
	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)

	var response errorResponse
	json.Unmarshal(buf.Bytes(), &response)

	// test response
	assert.Equal(t, "port is empty", response.Error)
}

// write test for request /_/yosoy/ping with h=127.0.0.1 parameter and p=8123 parameter, the request should return bad request 400 error and return JSON error about tcp connection issue
func TestHandlerPingWithHostnameAndPort(t *testing.T) {
	req, err := http.NewRequest("GET", "https://example.org/_/yosoy/ping?h=127.0.0.1&p=8123", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "*/*")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ping)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	result := rr.Result()
	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)

	var response errorResponse
	json.Unmarshal(buf.Bytes(), &response)

	// test response
	assert.Equal(t, "dial tcp 127.0.0.1:8123: connect: connection refused", response.Error)
}

// write test for request /_/yosoy/ping with h=127.0.0.1 parameter and p=8123 parameter, the request should return 200 ok and return JSON with message ping succeeded
func TestHandlerPingWithHostnameAndPortSuccess(t *testing.T) {

	// create tcp process to listen on port 8123
	listener, err := net.Listen("tcp", "localhost:8123")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	req, err := http.NewRequest("GET", "https://example.org/_/yosoy/ping?h=127.0.0.1&p=8123", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "*/*")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ping)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	result := rr.Result()
	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)

	var response successResponse
	json.Unmarshal(buf.Bytes(), &response)

	// test response
	assert.Equal(t, "ping succeeded", response.Message)
}

// write test for request /_/yosoy/ping?h=127.0.0.1&p=8123&n=qwq, the request should return 500 internal server error and return JSON with error "dial qwq: unknown network qwq"
func TestHandlerPingWithHostnameAndPortAndNetwork(t *testing.T) {

	// create tcp process to listen on port 8123
	listener, err := net.Listen("tcp", "localhost:8123")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	req, err := http.NewRequest("GET", "https://example.org/_/yosoy/ping?h=127.0.0.1&p=8123&n=qwq", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "*/*")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ping)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	result := rr.Result()
	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)
	var response errorResponse
	json.Unmarshal(buf.Bytes(), &response)
	// test response
	assert.Equal(t, "dial qwq: unknown network qwq", response.Error)
}

// write test for preflight HTTP Options request, verify that all headers are set
func TestHandlerPreflight(t *testing.T) {
	req, err := http.NewRequest("OPTIONS", "https://example.org/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "*/*")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(preflight)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	result := rr.Result()
	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)

	// test response
	assert.Equal(t, "*", result.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "*", result.Header.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "*", result.Header.Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "true", result.Header.Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "600", result.Header.Get("Access-Control-Max-Age"))
	assert.Equal(t, "*", result.Header.Get("Access-Control-Expose-Headers"))
}

// write test for request /_/yosoy/ping?h=127.0.0.1&p=8123&n=tcp&t=5, the request should return 200 ok and return JSON with message "ping succeeded"
func TestHandlerPingWithHostnameAndPortAndNetworkAndTimeout(t *testing.T) {

	// create tcp process to listen on port 8123
	listener, err := net.Listen("tcp", "localhost:8123")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	req, err := http.NewRequest("GET", "https://example.org/_/yosoy/ping?h=127.0.0.1&p=8123&n=tcp&t=5", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "*/*")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ping)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	result := rr.Result()
	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)
	var response successResponse
	json.Unmarshal(buf.Bytes(), &response)
	// test response
	assert.Equal(t, "ping succeeded", response.Message)
}

// write test for request /_/yosoy/ping?h=127.0.0.1&p=8123&n=tcp&t=invalid, the request should return 400 bad request and return JSON with error "timeout is invalid"
func TestHandlerPingWithHostnameAndPortAndNetworkAndTimeoutInvalid(t *testing.T) {

	// create tcp process to listen on port 8123
	listener, err := net.Listen("tcp", "localhost:8123")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	req, err := http.NewRequest("GET", "https://example.org/_/yosoy/ping?h=127.0.0.1&p=8123&n=tcp&t=invalid", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "*/*")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ping)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	result := rr.Result()
	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)
	var response errorResponse
	json.Unmarshal(buf.Bytes(), &response)
	// test response
	assert.Equal(t, "timeout is invalid", response.Error)
}
