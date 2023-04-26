package controller

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

func getPasswordHash(password string) (string, error) {
	log.Println("in getPasswordHash")
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 6)
	if err == nil {
		return string(hashed), nil
	}

	return "", err
}

func validatePassword(hashedPassword string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
