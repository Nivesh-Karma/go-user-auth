package models

import "time"

type TokenModel struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Expire       time.Time `json:"expire"`
	UserData     UserModel `json:"user"`
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

type UserModel struct {
	FirstName        string   `json:"first_name"`
	LastName         string   `json:"last_name"`
	Premium          bool     `json:"premium"`
	AddRatios        []string `json:"add_ratios"`
	PeerRatios       []string `json:"peer_ratios"`
	HistoricalRatios []string `json:"historical_ratios"`
}
