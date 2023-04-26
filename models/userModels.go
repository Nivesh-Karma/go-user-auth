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
	Username          string      `form:"username" binding:"required,email"`
	Password          string      `form:"password" binding:"required"`
	FirstName         string      `form:"first_name" binding:"required"`
	LastName          string      `form:"last_name" binding:"required"`
	SecurityQuestion1 ChoicesEnum `form:"security_question_1" binding:"required"`
	SecurityAnswer1   string      `form:"security_answer_1"`
}

type UserResponse struct {
	Username  string `form:"username"`
	FirstName string `form:"first_name"`
	LastName  string `form:"last_name"`
	Admin     bool   `form:"admin"`
	Active    bool   `form:"active"`
	Premium   bool   `form:"premium"`
}

type LoginRequest struct {
	Username string `form:"username" binding:"required,email"`
	Password string `form:"password" binding:"required"`
}

type ResetPassword struct {
	Username          string      `form:"username" binding:"required,email"`
	NewPassword       string      `form:"new_password" binding:"required"`
	SecurityQuestion1 ChoicesEnum `form:"security_question_1" binding:"required"`
	SecurityAnswer1   string      `form:"security_answer_1" binding:"required"`
}

type UnlockAccount struct {
	Username          string      `form:"username" binding:"required,email"`
	SecurityQuestion1 ChoicesEnum `form:"security_question_1" binding:"required"`
	SecurityAnswer1   string      `form:"security_answer_1" binding:"required"`
}

type AdminUpdates struct {
	Username      string `form:"username" binding:"required,email"`
	IsNewPassword bool   `form:"is_new_password" binding:"required"`
	NewPassword   string `form:"password" binding:"required"`
	Activate      bool   `form:"activate"`
	Deactivate    bool   `form:"deactivate"`
	Lock          bool   `form:"lock"`
	Unlock        bool   `form:"unlock"`
	Promote       bool   `form:"promote"`
	Demote        bool   `form:"demote"`
}
