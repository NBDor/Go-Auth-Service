//go:build database

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
	// Configure environment for database tests
	os.Setenv("DB_HOST", os.Getenv("DB_HOST"))
	os.Setenv("DB_PORT", os.Getenv("DB_PORT"))
	os.Setenv("DB_USER", os.Getenv("DB_USER"))
	os.Setenv("DB_PASSWORD", os.Getenv("DB_PASSWORD"))
	os.Setenv("DB_NAME", os.Getenv("DB_NAME"))
	os.Setenv("DB_SSLMODE", os.Getenv("DB_SSLMODE"))
	os.Setenv("JWT_SECRET", "test-secret-key")
}

func TestDBHealthEndpoint(t *testing.T) {
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

func TestDBAuthentication(t *testing.T) {
	router, _ := server.SetupRouter()
	
	// Test authentication with valid credentials for the database
	form := url.Values{}
	form.Add("username", "admin")  // This user is created in the database
	form.Add("password", "admin123")
	
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
	form.Add("username", "admin")
	form.Add("password", "wrongpassword")
	
	req = httptest.NewRequest("POST", "/auth/login", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
