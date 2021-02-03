package main

import (
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

	req, err := http.NewRequest("GET", "https://example.org/sample/path", nil)
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

	var response response
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.Equal(t, 1, response.Counter)
	assert.Equal(t, "example.org", response.Host)
	assert.NotEmpty(t, response.EnvVariables)
	assert.NotEmpty(t, response.Files[".gitignore"])
}
