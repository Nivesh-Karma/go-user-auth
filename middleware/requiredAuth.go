package middleware

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func VerifyJWTToken(tokenString string, scope string) (string, error) {
	var secretKey = os.Getenv("SECRET_KEY")
	log.Println(tokenString)
	credentialsException := errors.New("could not validate credentials")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", credentialsException
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", credentialsException
	}
	if claims["scope"] != scope {
		return "", errors.New("invalid scope for the token")
	}
	username, ok := claims["sub"].(string)
	if !ok {
		return "", credentialsException
	}
	return username, nil
}

func RequireAuth(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	splitToken := strings.Split(tokenString, "Bearer")
	if len(splitToken) != 2 {
		// Error: Bearer token not in proper format
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unable to find Bearer token"})
	}

	tokenString = strings.TrimSpace(splitToken[1])

	username, err := VerifyJWTToken(tokenString, "auth")
	if err != nil {
		log.Println("RequireAuth", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to verify token"})
	}
	c.Set("username", username)
	c.Next()
}
