package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	os.Setenv("YOSOY_SHOW_ENVS", "true")
	os.Setenv("YOSOY_SHOW_FILES", ".gitignore")

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
