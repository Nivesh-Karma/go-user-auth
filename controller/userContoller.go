package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Nivesh-Karma/go-user-admin/models"
	"github.com/gin-gonic/gin"
)

func CreateNewUser(c *gin.Context) {
	log.Println("CreateNewUser invoked")
	var user models.UserRequest
	if err := c.Bind(&user); err != nil {
		log.Println("error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	if _, ok := findUser(user.Username); ok {
		log.Printf("User %s already exists", user.Username)
		c.JSON(http.StatusForbidden, gin.H{"error": "user already exists!"})
		return
	}
	hashed, err := getPasswordHash(user.Password)
	if err != nil {
		log.Println("error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating the user, try again later!"})
		return
	}
	user.Password = hashed

	if userStatus := createUser(&user, "email"); userStatus {
		c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("user %s created successfully", user.Username)})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating the user, try again later!"})
	}
}

func Login(c *gin.Context) {
	log.Println("CreateNewUser invoked")
	var userLogin models.LoginRequest
	if err := c.Bind(&userLogin); err != nil {
		log.Println("error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}
	user, ok := findUser(userLogin.Username)
	if !ok {
		log.Printf("User %s does not exist", userLogin.Username)
		c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("user %s doesnot exist", userLogin.Username)})
		return
	}
	if user.Locked {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "account is locked due to multiple failed attempts."})
	}
	if !user.Active {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "account has been deactivated. Please contact support."})
	}
	isValid := validatePassword(user.Password, userLogin.Password)
	if !isValid {
		go updateFailedCounter(user)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect username or password"})
		return
	} else {
		go resetFailedCount(user)
	}
	if token, err := createJWTToken(user.Username); err != nil {
		log.Println("error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	} else {
		c.IndentedJSON(http.StatusOK, token)
	}
}

func Validate(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		c.AbortWithStatus(http.StatusNotAcceptable)
	}
	_, ok = findUser(username.(string))
	if !ok {
		log.Printf("User %s does not exist", username)
		c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("user %s doesnot exist", username)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "valid token"})
}

func ValidateAdmin(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		c.AbortWithStatus(http.StatusNotAcceptable)
	}
	user, ok := isAdmin(username.(string))
	if !ok {
		log.Printf("User %s does not exist", username)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("user %s doesnot exist", username)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"admin": user})
}

func isAdmin(username string) (bool, bool) {
	user, ok := findUser(username)
	return user.Active, ok
}

func ResetPassword(c *gin.Context) {
	var userLogin models.ResetPassword
	if err := c.ShouldBindJSON(&userLogin); err != nil {
		log.Println("error: ", err)
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Failed to read body"})
	}
	user, ok := findUser(userLogin.Username)
	if !ok {
		log.Printf("User %s does not exist", userLogin.Username)
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": fmt.Sprintf("user %s doesnot exist", userLogin.Username)})
	}
	if user.UserSource != "email" {
		c.AbortWithStatusJSON(http.StatusNotAcceptable,
			gin.H{"error": fmt.Sprintf("Please use %s authentication. Password reset not allowed.",
				user.UserSource)})
	}
	if user.SecurityQuestion1 != string(userLogin.SecurityQuestion1) || user.SecurityAnswer1 != string(userLogin.SecurityAnswer1) {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Security question or answer does not match"})
	}
	hashed, err := getPasswordHash(userLogin.NewPassword)
	if err != nil {
		log.Println("error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating the user, try again later!"})
		return
	}
	user.Password = hashed
	user.FailedCount = 0
	user.Locked = false
	updateUser(user)
	c.JSON(http.StatusAccepted, gin.H{"message": "password reset successful"})
}

func UnlockAccount(c *gin.Context) {
	var userLogin models.UnlockAccount
	if err := c.ShouldBindJSON(&userLogin); err != nil {
		log.Println("error: ", err)
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Failed to read body"})
	}
	user, ok := findUser(userLogin.Username)
	if !ok {
		log.Printf("User %s does not exist", userLogin.Username)
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": fmt.Sprintf("user %s doesnot exist", userLogin.Username)})
	}
	if user.UserSource != "email" {
		c.AbortWithStatusJSON(http.StatusNotAcceptable,
			gin.H{"error": fmt.Sprintf("Please use %s authentication. Password reset not allowed.",
				user.UserSource)})
	}
	if user.SecurityQuestion1 != string(userLogin.SecurityQuestion1) || user.SecurityAnswer1 != string(userLogin.SecurityAnswer1) {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Security question or answer does not match"})
	}
	user.FailedCount = 0
	user.Locked = false
	updateUser(user)
	c.JSON(http.StatusAccepted, gin.H{"message": "Unlocked the account"})
}

func AdminUpdates(c *gin.Context) {
	adminUsername, ok := c.Get("username")
	if !ok {
		c.AbortWithStatus(http.StatusNotAcceptable)
	}
	isAdminUser, _ := isAdmin(adminUsername.(string))
	if !isAdminUser {
		c.AbortWithStatus(http.StatusNotAcceptable)
	}
	adminRequest := models.AdminUpdates{}
	if err := c.ShouldBindJSON(&adminRequest); err != nil {
		log.Println("error: ", err)
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Failed to read body"})
	}
	user, ok := findUser(adminRequest.Username)
	if !ok {
		log.Printf("User %s not found\n", adminRequest.Username)
		c.AbortWithStatus(http.StatusNotAcceptable)
	}
	if adminRequest.IsNewPassword {
		user.Password, _ = getPasswordHash(adminRequest.NewPassword)
	}
	if adminRequest.Activate {
		user.Active = true
		user.FailedCount = 0
		user.Locked = false
	}
	if adminRequest.Deactivate {
		user.Active = false
	}
	if adminRequest.Unlock {
		user.Locked = false
		user.FailedCount = 0
	}
	if adminRequest.Lock {
		user.Locked = true
	}
	updateUser(user)
	c.JSON(http.StatusAccepted, gin.H{"message": "updated all"})
}
