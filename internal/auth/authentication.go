package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kwekkwekpatu/chirpy/internal/util"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	util.InfoLogger.Printf("Generating new JWT for %s", userID.String())
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedJWT, err := token.SignedString([]byte(tokenSecret))

	util.InfoLogger.Printf("Succesfully generated JWT for %s", userID.String())
	return signedJWT, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	util.InfoLogger.Printf("Validating JWT")

	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(tokenSecret), nil
		})
	if err != nil {
		util.ErrorLogger.Printf("Failed to validate JWT with error: %s", err.Error())
		return uuid.UUID{}, err
	}

	if !token.Valid {
		return uuid.UUID{}, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("invalid claims type")
	}

	userIDString, err := claims.GetSubject()
	if err != nil {
		util.ErrorLogger.Printf("Failed to retriev UserID with error: %s", err.Error())
		return uuid.UUID{}, err
	}

	util.InfoLogger.Printf("Succesfully validated JWT for %s", userIDString)
	return uuid.Parse(userIDString)
}
