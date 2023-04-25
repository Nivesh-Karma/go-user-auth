package models

import (
	"time"
)

type ChoicesEnum string

const (
	Q1 ChoicesEnum = "In what city were you born?"
	Q2 ChoicesEnum = "What is your mother's maiden name?"
	Q3 ChoicesEnum = "What was the make of your first car?"
	Q4 ChoicesEnum = "What is your favorite teacher's name?"
	Q5 ChoicesEnum = "What is the name of your favorite pet?"
	Q6 ChoicesEnum = "Is the user logged from google?"
)

type Users struct {
	UserID            int    `gorm:"primaryKey"`
	Username          string `gorm:"uniqueIndex"`
	Password          string
	FirstName         string
	LastName          string
	Admin             bool `gorm:"default:false"`
	Active            bool `gorm:"default:true"`
	PremiumUser       bool `gorm:"default:false"`
	PremiumExpiryDate time.Time
	Locked            bool      `gorm:"default:false"`
	FailedCount       int       `gorm:"default:0"`
	SecurityQuestion1 string    `gorm:"default:'';column:security_question_1"`
	SecurityAnswer1   string    `gorm:"default:'';column:security_answer_1"`
	UserSource        string    `gorm:"default:'email'"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`
}

func (Users) TableName() string {
	return "users"
}

type UserRequest struct {
	Username          string      `json:"username" binding:"required,email"`
	Password          string      `json:"password" binding:"required"`
	FirstName         string      `json:"first_name" binding:"required"`
	LastName          string      `json:"last_name" binding:"required"`
	SecurityQuestion1 ChoicesEnum `json:"security_question_1" binding:"required"`
	SecurityAnswer1   string      `json:"security_answer_1"`
}

type UserResponse struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Admin     bool   `json:"admin"`
	Active    bool   `json:"active"`
	Premium   bool   `json:"premium"`
}

type LoginRequest struct {
	Username string `form:"username" binding:"required,email"`
	Password string `form:"password" binding:"required"`
}

type ResetPassword struct {
	Username          string      `json:"username" binding:"required,email"`
	NewPassword       string      `json:"new_password" binding:"required"`
	SecurityQuestion1 ChoicesEnum `json:"security_question_1" binding:"required"`
	SecurityAnswer1   string      `json:"security_answer_1" binding:"required"`
}

type UnlockAccount struct {
	Username          string      `json:"username" binding:"required,email"`
	SecurityQuestion1 ChoicesEnum `json:"security_question_1" binding:"required"`
	SecurityAnswer1   string      `json:"security_answer_1" binding:"required"`
}

type AdminUpdates struct {
	IsNewPassword bool   `json:"is_new_password" binding:"required"`
	NewPassword   string `json:"password" binding:"required"`
	Username      string `json:"username" binding:"required,email"`
	Activate      bool   `json:"activate"`
	Deactivate    bool   `json:"deactivate"`
	Lock          bool   `json:"lock"`
	Unlock        bool   `json:"unlock"`
}
