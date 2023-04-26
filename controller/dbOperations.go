package controller

import (
	"log"

	"github.com/Nivesh-Karma/go-user-admin/config"
	"github.com/Nivesh-Karma/go-user-admin/models"
)

func findUser(email string) (*models.Users, bool) {
	log.Println("in findUser")
	user := models.Users{}
	if result := config.DB.Where("username ILIKE ?", email).First(&user); result.Error != nil {
		log.Println("Error:", result.Error)
		return &user, false
	}
	return &user, true
}

func createUser(user *models.UserRequest, userSource string) bool {
	log.Println("in createUser")
	userDB := models.Users{
		Username:          user.Username,
		Password:          user.Password,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		UserSource:        userSource,
		SecurityQuestion1: string(user.SecurityQuestion1),
		SecurityAnswer1:   user.SecurityAnswer1,
	}
	if result := config.DB.Create(&userDB); result.Error != nil {
		return false
	}
	return true
}

func updateFailedCounter(user *models.Users) {
	user.FailedCount += 1
	if user.FailedCount > 5 {
		user.Locked = true
	}
	config.DB.Save(user)
}

func resetFailedCount(user *models.Users) {
	if user.FailedCount < 1 && !user.Locked {
		return
	}
	user.FailedCount = 0
	user.Locked = false
	config.DB.Save(user)
}

func updateUser(user *models.Users) {
	config.DB.Save(user)
}
