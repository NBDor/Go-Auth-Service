//go:build !database

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/NBDor/Go-Auth-Service/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestMemoryTokenRevocation(t *testing.T) {
	router, _ := server.SetupRouter()
	
	// 1. Login to get a token
	form := url.Values{}
	form.Add("username", "testuser")
	form.Add("password", "password123")
	
	loginReq := httptest.NewRequest("POST", "/auth/login", strings.NewReader(form.Encode()))
	loginReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	loginW := httptest.NewRecorder()
	
	router.ServeHTTP(loginW, loginReq)
	
	assert.Equal(t, http.StatusOK, loginW.Code)
	
	var loginResponse map[string]interface{}
	err := json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
	assert.NoError(t, err)
	
	token, ok := loginResponse["token"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, token)
	
	// 2. Access protected endpoint with the token
	meReq := httptest.NewRequest("GET", "/auth/me", nil)
	meReq.Header.Add("Authorization", "Bearer "+token)
	meW := httptest.NewRecorder()
	
	router.ServeHTTP(meW, meReq)
	
	assert.Equal(t, http.StatusOK, meW.Code)
	
	var meResponse map[string]interface{}
	err = json.Unmarshal(meW.Body.Bytes(), &meResponse)
	assert.NoError(t, err)
	assert.Contains(t, meResponse, "user")
	
	// 3. Revoke the token
	logoutReq := httptest.NewRequest("POST", "/auth/logout", nil)
	logoutReq.Header.Add("Authorization", "Bearer "+token)
	logoutW := httptest.NewRecorder()
	
	router.ServeHTTP(logoutW, logoutReq)
	
	assert.Equal(t, http.StatusOK, logoutW.Code)
	
	// 4. Try to access protected endpoint with revoked token
	meReq2 := httptest.NewRequest("GET", "/auth/me", nil)
	meReq2.Header.Add("Authorization", "Bearer "+token)
	meW2 := httptest.NewRecorder()
	
	router.ServeHTTP(meW2, meReq2)
	
	// Should be unauthorized with revoked token
	assert.Equal(t, http.StatusUnauthorized, meW2.Code)
}

func TestMemoryProtectedEndpoint(t *testing.T) {
	router, _ := server.SetupRouter()
	
	// Try accessing protected endpoint without token
	req := httptest.NewRequest("GET", "/auth/me", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	// Try with invalid token
	req = httptest.NewRequest("GET", "/auth/me", nil)
	req.Header.Add("Authorization", "Bearer invalid.token.here")
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
