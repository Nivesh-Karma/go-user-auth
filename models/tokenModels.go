package models

import "time"

type TokenModel struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Expire       time.Time `json:"expire"`
}

type RefreshTokenModel struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	Expire      time.Time `json:"expire"`
}

type TokenRequest struct {
	AccessToken string
	Scope       string
	Expire      time.Time
	Err         error
}
