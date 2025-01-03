package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/kwekkwekpatu/chirpy/internal/util"
)

func GetAPIKey(headers http.Header) (string, error) {
	util.InfoLogger.Println("Getting APIKey")
	authorizationHeader := headers.Get("Authorization")
	authorizationFields := strings.Fields(authorizationHeader)
	if len(authorizationFields) != 2 {
		util.ErrorLogger.Println("Failed to get APIKey. Invalid authorization header format")
		return "", fmt.Errorf("Invalid authorization header format")
	}

	if authorizationFields[0] != "ApiKey" {
		util.ErrorLogger.Println("Failed to get APIKey. First part is not 'ApiKey'")
		return "", fmt.Errorf("Authorization header must start with 'ApiKey'")
	}

	util.InfoLogger.Println("Successfully retrieved APIKey")
	return authorizationFields[1], nil
}
