package jwt

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

func GenerateToken(id string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": id,
	})

	tokenString, err := token.SignedString([]byte("young"))
	if err != nil {
		fmt.Println("Error generating token:", err)
		return ""
	}
	return tokenString
}

func AuthenticateToken(tokenString string) (string, error) {
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("young"), nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["id"].(string), nil
	}
	return "", nil
}

func GetIdFromToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, nil)
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims["id"].(string), nil
	}
	return "", err
}
