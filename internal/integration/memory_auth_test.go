//go:build !database

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/NBDor/Go-Auth-Service/internal/server"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Ensure we use in-memory storage by setting invalid DB credentials
	os.Setenv("DB_HOST", "nonexistent-host")
	os.Setenv("JWT_SECRET", "test-secret-key")
}

func TestMemoryHealthEndpoint(t *testing.T) {
	router, _ := server.SetupRouter()
	
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
	assert.Contains(t, response, "providers")
}

func TestMemoryAuthentication(t *testing.T) {
	router, _ := server.SetupRouter()
	
	// Test authentication with valid credentials for the in-memory store
	form := url.Values{}
	form.Add("username", "testuser")  // This user is created in the in-memory store
	form.Add("password", "password123")
	
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.NoError(t, err)
	assert.Contains(t, response, "token")
	assert.Contains(t, response, "user")
	
	// Test with invalid credentials
	form = url.Values{}
	form.Add("username", "testuser")
	form.Add("password", "wrongpassword")
	
	req = httptest.NewRequest("POST", "/auth/login", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
