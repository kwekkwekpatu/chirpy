package auth_test

import (
	"net/http"
	"testing"

	"github.com/kwekkwekpatu/chirpy/internal/auth"
)

func TestGetBearerToken(t *testing.T) {
	testCases := []struct {
		name          string
		headers       http.Header
		expectedToken string
		expectedError bool
	}{
		{
			name: "valid bearer token",
			headers: http.Header{
				"Authorization": []string{"Bearer abc123.def456.ghi789"},
			},
			expectedToken: "abc123.def456.ghi789",
			expectedError: false,
		},
		{
			name:          "missing authorization header",
			headers:       http.Header{},
			expectedToken: "",
			expectedError: true,
		},
		{
			name: "missing bearer prefix",
			headers: http.Header{
				"Authorization": []string{"abc123.def456.ghi789"},
			},
			expectedToken: "",
			expectedError: true,
		},
		{
			name: "empty bearer token",
			headers: http.Header{
				"Authorization": []string{"Bearer "},
			},
			expectedToken: "",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := auth.GetBearerToken(tc.headers)

			if tc.expectedError && err == nil {
				t.Error("expected error but got none")
			}
			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if token != tc.expectedToken {
				t.Errorf("expected token %q, got %q", tc.expectedToken, token)
			}
		})
	}
}
