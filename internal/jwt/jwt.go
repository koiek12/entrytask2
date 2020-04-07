package jwt

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Issue JWT token
type TokenIssuer struct {
	key        string
	expireTime time.Duration
}

func NewTokenIssuer(key string, expireTime time.Duration) *TokenIssuer {
	return &TokenIssuer{
		key:        key,
		expireTime: time.Minute * expireTime,
	}
}

type TokenExpiredError struct{}

func (e TokenExpiredError) Error() string {
	return "token has expired"
}

// generate JWT token.
func (issuer *TokenIssuer) GenerateToken(id string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(issuer.expireTime).Unix(),
	})

	tokenString, err := token.SignedString([]byte("young"))
	if err != nil {
		return ""
	}
	return tokenString
}

// authenticate JWT token, on success return owner id of token, on fail return empty string
func (issuer *TokenIssuer) AuthenticateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("young"), nil
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

// extract ID from JWT token claim
func GetIdFromToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, nil)
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims["id"].(string), nil
	}
	return "", err
}
