package main

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/http2"
)

func TestHTTP2Protocol(t *testing.T) {
       // Create a test server with HTTP/2 support
       server := httptest.NewUnstartedServer(http.HandlerFunc(handler))
       server.EnableHTTP2 = true
       server.StartTLS()
       defer server.Close()

       // Create HTTP/2 client
       client := &http.Client{
               Transport: &http2.Transport{
                       TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
               },
       }

       resp, err := client.Get(server.URL + "/test")
       assert.NoError(t, err)
       defer resp.Body.Close()

       assert.Equal(t, http.StatusOK, resp.StatusCode)
       assert.Equal(t, "HTTP/2.0", resp.Proto)

       var response response
       err = json.NewDecoder(resp.Body).Decode(&response)
       assert.NoError(t, err)
       assert.Equal(t, "HTTP/2.0", response.Proto)
}

func TestProtocolVersionReporting(t *testing.T) {
	testCases := []struct {
		name        string
		proto       string
		protoMajor  int
		protoMinor  int
		expected    string
	}{
		{"HTTP/1.0", "HTTP/1.0", 1, 0, "HTTP/1.0"},
		{"HTTP/1.1", "HTTP/1.1", 1, 1, "HTTP/1.1"},
		{"HTTP/2.0", "HTTP/2.0", 2, 0, "HTTP/2.0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/test", nil)
			assert.NoError(t, err)
	
			req.Proto = tc.proto
			req.ProtoMajor = tc.protoMajor
			req.ProtoMinor = tc.protoMinor

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
	
			var response response
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, response.Proto)
		})
	}
}
