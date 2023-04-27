package controller

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Nivesh-Karma/go-user-admin/models"
	"github.com/golang-jwt/jwt/v5"
)

func createJWTToken(username string) (*models.TokenModel, error) {
	wg := sync.WaitGroup{}
	token := &models.TokenModel{TokenType: "Bearer"}
	dataChan := make(chan *models.TokenRequest, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		tokenObject := createAuthToken(username, "auth")
		dataChan <- tokenObject
	}()
	go func() {
		defer wg.Done()
		tokenObject := createAuthToken(username, "refresh")
		dataChan <- tokenObject
	}()
	wg.Wait()
	close(dataChan)
	for n := range dataChan {
		if n.Err != nil {
			return nil, n.Err
		}
		if n.Scope == "auth" {
			token.AccessToken = n.AccessToken
			token.Expire = n.Expire
		} else {
			token.RefreshToken = n.AccessToken
		}
	}
	return token, nil
}

func CreateRefreshToken(username string) *models.TokenRequest {
	return createAuthToken(username, "refresh")
}

func createAuthToken(username, scope string) *models.TokenRequest {
	secretKey := os.Getenv("SECRET_KEY")
	expireMinutes := os.Getenv("ACCESS_TOKEN_EXPIRE_MINUTES")
	duration, _ := strconv.Atoi(expireMinutes)
	if scope != "auth" {
		duration = 60 * 24 * 30
	}
	tokenRequest := models.TokenRequest{Scope: scope}
	expTime := time.Now().Add(time.Minute * time.Duration(duration)).Unix()
	toEncode := jwt.MapClaims{
		"sub":   username,
		"scope": scope,
		"exp":   expTime,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, toEncode)
	var err error
	tokenRequest.AccessToken, err = token.SignedString([]byte(secretKey))
	if err != nil {
		tokenRequest.Err = err
		return &tokenRequest
	}
	tokenRequest.Expire = time.Unix(toEncode["exp"].(int64), 0)
	return &tokenRequest
}

func refreshAuthToken(username string) (*models.RefreshTokenModel, error) {
	tokenRequest := createAuthToken(username, "auth")
	if tokenRequest.Err != nil {
		return nil, tokenRequest.Err
	}
	token := &models.RefreshTokenModel{
		AccessToken: tokenRequest.AccessToken,
		Expire:      tokenRequest.Expire,
		TokenType:   "Bearer",
	}
	return token, nil
}
