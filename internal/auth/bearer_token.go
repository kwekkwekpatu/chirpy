package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/kwekkwekpatu/chirpy/internal/util"
)

const (
	RandomBytesCount = 32
)

func GetBearerToken(headers http.Header) (string, error) {
	util.InfoLogger.Println("Getting bearer token")
	authorizationHeader := headers.Get("Authorization")
	if authorizationHeader == "" {
		util.ErrorLogger.Println("Failed to get bearer token. Authorization Header is empty.")
		return "", fmt.Errorf("Authorization header is empty")
	}

	authorizationFields := strings.Fields(authorizationHeader)
	if len(authorizationFields) != 2 {
		util.ErrorLogger.Println("Failed to get bearer token. Invalid authorization header format")
		return "", fmt.Errorf("Invalid authorization header format")
	}

	if authorizationFields[0] != "Bearer" {
		util.ErrorLogger.Println("Failed to get bearer token. First part is not 'Bearer'")
		return "", fmt.Errorf("Authorization header must start with 'Bearer'")
	}

	util.InfoLogger.Println("Successfully retrieved bearer token")
	return authorizationFields[1], nil
}

func MakeRefreshToken() (string, error) {
	util.InfoLogger.Println("Making refresh token")
	tokenbytes := make([]byte, RandomBytesCount)
	_, err := rand.Read(tokenbytes)
	if err != nil {
		return "", fmt.Errorf("Failed to make random bytes")
	}

	util.InfoLogger.Println("Succesfully made token")
	return hex.EncodeToString(tokenbytes), nil
}
