package mware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIPMiddleware(t *testing.T) {
	// Mock handler to be wrapped by middleware
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Define test cases
	tests := []struct {
		name           string
		trustedSubnet  string
		xRealIP        string
		expectedStatus int
	}{
		{
			name:           "Valid trusted IP",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.1.10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Untrusted IP",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.2.10",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid IP format in header",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "invalid-ip",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid subnet format",
			trustedSubnet:  "invalid-subnet",
			xRealIP:        "192.168.1.10",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "No trusted subnet defined",
			trustedSubnet:  "",
			xRealIP:        "192.168.2.10",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Wrap the handler with GetIPMiddleware
			handler := GetIPMiddleware(mockHandler, tc.trustedSubnet)

			// Create a request with the X-Real-IP header
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-Real-IP", tc.xRealIP)

			// Record the response
			recorder := httptest.NewRecorder()

			// Serve the request
			handler.ServeHTTP(recorder, req)

			// Assert the response status code
			assert.Equal(t, tc.expectedStatus, recorder.Code)
		})
	}
}
