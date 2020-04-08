package jwt

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// TokenIssuer issues JWT token for user validation. Tokens have expiration time for security
type TokenIssuer struct {
	key        string
	expireTime time.Duration
}

// NewTokenIssuer create and return TokenIssuer with secret key and expiration time.
func NewTokenIssuer(key string, expireTime time.Duration) *TokenIssuer {
	return &TokenIssuer{
		key:        key,
		expireTime: expireTime,
	}
}

// TokenExpiredError occurs when token exp field is before current time
type TokenExpiredError struct{}

func (e TokenExpiredError) Error() string {
	return "token has expired"
}

// GenerateToken generate JWT token with secret key and claims of user's id, expiration time and issue time.
// On succes, returns generated token. On fail return empty string
func (issuer *TokenIssuer) GenerateToken(id string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(issuer.expireTime).Unix(),
	})

	tokenString, err := token.SignedString([]byte(issuer.key))
	if err != nil {
		return ""
	}
	return tokenString
}

// AuthenticateToken check given JWT token's signature with it's secret key.
// It also check expiration date and issue date. Any one of verification fails, returns error.
// On success return owner id of token.
func (issuer *TokenIssuer) AuthenticateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(issuer.key), nil
	})
	if err != nil {
		return "", err
	}
	// check expire date and issue date
	err = token.Claims.Valid()
	if err != nil {
		return "", nil
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["id"].(string), nil
	}
	return "", nil
}

// GetIDFromToken extract id claim from JWT token
func GetIDFromToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, nil)
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims["id"].(string), nil
	}
	return "", err
}
