package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	TokenTypeAccess     TokenType = "chirpy-access"
	AuthorizationPrefix string    = "Authorization"
	TokenPrefix         string    = "Bearer"
	APIKeyPrefix        string    = "ApiKey"
)

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func CheckPasswordHash(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	signingKey := []byte(tokenSecret)
	claims := jwt.RegisteredClaims{
		Issuer:    string(TokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(signingKey)
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claimsStruct, func(*jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	id, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return id, nil
}

func MakeRefreshToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

/*
 * Wrapped function for extracting values from header keys
 * Requires 2 error in errorMsgs:
 * 		errorMsgs[0] = error when there is no value for headerKey in the header
 *		errorMsgs[1] = error when the header prefix is missing or doesn't match the expected one
 */
func getHeaderValue(headers http.Header, headerKey string, headerPrefix string, errorMsgs []error) (string, error) {
	if len(errorMsgs) < 2 {
		return "", fmt.Errorf("getting headers from %s requires 2 error responses for proper handling", headerKey)
	}

	headerValue := headers.Get(headerKey)
	if headerValue == "" {
		return "", errorMsgs[0]
	}

	headerParts := strings.Split(headerValue, " ")
	if len(headerParts) < 2 || headerParts[0] != headerPrefix {
		return "", errorMsgs[1]
	}

	return headerParts[1], nil
}

func GetBearerToken(headers http.Header) (string, error) {
	var errNoAuthHeaderIncluded = errors.New("no authorization header included in request")
	var errMalformedHeader = errors.New("malformed authorization header")

	errMsgs := []error{errNoAuthHeaderIncluded, errMalformedHeader}
	/*
		authHeader := headers.Get("Authorization")
		if authHeader == "" {
			return "", ErrNoAuthHeaderIncluded
		}

		authParts := strings.Split(authHeader, " ")
		if len(authParts) < 2 || authParts[0] != TokenPrefix {
			return "", errors.New("malformed authorization header")
		}

		return authParts[1], nil
	*/
	return getHeaderValue(headers, AuthorizationPrefix, TokenPrefix, errMsgs)
}

func GetAPIKey(headers http.Header) (string, error) {
	var errNoAPIKeyIncluded = errors.New("no API key included in request")
	var errMalformedHeader = errors.New("malformed API key")
	errorMsgs := []error{errNoAPIKeyIncluded, errMalformedHeader}
	/*
		keyHeader := headers.Get("Authorization")
		if keyHeader == "" {
			return "", ErrNoAPIKeyIncluded
		}

		keyParts := strings.Split(keyHeader, " ")
		if len(keyParts) < 2 || keyParts[0] != APIPrefix {
			return "", errors.New("malformed API key")
		}

		return keyParts[1], nil
	*/
	return getHeaderValue(headers, AuthorizationPrefix, APIKeyPrefix, errorMsgs)
}
