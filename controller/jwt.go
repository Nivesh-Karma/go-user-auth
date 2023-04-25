package controller

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Nivesh-Karma/go-user-admin/middleware"
	"github.com/Nivesh-Karma/go-user-admin/models"
	"github.com/golang-jwt/jwt/v5"
)

func createJWTToken(username string) (*models.TokenModel, error) {

	encodedJWT, expire, err := createAuthToken(username, "auth")
	if err != nil {
		return nil, err
	}
	refreshJWT, _, err := createAuthToken(username, "refresh")
	if err != nil {
		return nil, err
	}
	token := &models.TokenModel{
		AccessToken:  encodedJWT,
		Expire:       expire,
		RefreshToken: refreshJWT,
		TokenType:    "Bearer",
	}
	return token, nil
}

func CreateRefreshToken(username string) (string, time.Time, error) {
	return createAuthToken(username, "refresh")
}

func createAuthToken(username, scope string) (string, time.Time, error) {
	secretKey := os.Getenv("SECRET_KEY")
	expireMinutes := os.Getenv("ACCESS_TOKEN_EXPIRE_MINUTES")
	duration, _ := strconv.Atoi(expireMinutes)
	if scope != "auth" {
		duration = 60 * 24 * 30
	}
	expTime := time.Now().Add(time.Minute * time.Duration(duration)).Unix()
	log.Println("expTime=", expTime)
	toEncode := jwt.MapClaims{
		"sub":   username,
		"scope": scope,
		"exp":   expTime,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, toEncode)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", time.Time{}, err
	}
	expire := time.Unix(toEncode["exp"].(int64), 0)
	return tokenString, expire, nil
}

func RefreshAuthToken(refreshToken string) (*models.RefreshTokenModel, error) {
	username, err := middleware.VerifyJWTToken(refreshToken, "refresh")
	if err != nil {
		return nil, err
	}
	encodedJWT, expire, err := createAuthToken(username, "auth")
	if err != nil {
		return nil, err
	}
	token := &models.RefreshTokenModel{
		AccessToken: encodedJWT,
		Expire:      expire,
		TokenType:   "Bearer",
	}
	return token, nil
}
