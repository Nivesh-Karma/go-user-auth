package controller

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Nivesh-Karma/go-user-admin/models"
	"github.com/gin-gonic/gin"
)

func CreateNewUser(c *gin.Context) {
	var user models.UserRequest
	if err := c.Bind(&user); err != nil {
		log.Println("error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}
	user.Username = strings.ToLower(user.Username)
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
	hashed, err = getPasswordHash(strings.ToLower(user.SecurityAnswer1))
	if err != nil {
		log.Println("error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating the user, try again later!"})
		return
	}
	user.SecurityAnswer1 = hashed

	if userStatus := createUser(&user, "email"); userStatus {
		c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("user %s created successfully", user.Username)})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating the user, try again later!"})
	}
}

func Login(c *gin.Context) {
	var userLogin models.LoginRequest
	if err := c.Bind(&userLogin); err != nil {
		log.Println("error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}
	userLogin.Username = strings.ToLower(userLogin.Username)
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
		if user.FailedCount > 0 || user.Locked {
			go resetFailedCount(user)
		}
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
	user, ok := findUser(username.(string))
	if !ok {
		log.Printf("User %s does not exist", username)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("user %s doesnot exist", username)})
		return
	}
	userResponse := models.UserResponse{
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Admin:     user.Admin,
		Active:    user.Active,
		Premium:   user.PremiumUser,
	}
	c.IndentedJSON(http.StatusOK, userResponse)
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
	if err := c.Bind(&userLogin); err != nil {
		log.Println("error: ", err)
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Failed to read body"})
		return
	}
	user, ok := findUser(userLogin.Username)
	if !ok {
		log.Printf("User %s does not exist", userLogin.Username)
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": fmt.Sprintf("user %s doesnot exist", userLogin.Username)})
		return
	}
	if user.UserSource != "email" {
		c.AbortWithStatusJSON(http.StatusNotAcceptable,
			gin.H{"error": fmt.Sprintf("Please use %s authentication. Password reset not allowed.",
				user.UserSource)})
		return
	}
	if user.SecurityQuestion1 != string(userLogin.SecurityQuestion1) ||
		validatePassword(user.SecurityAnswer1, strings.ToLower(string(userLogin.SecurityAnswer1))) {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Security question or answer does not match"})
		return
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
	if adminRequest.Promote {
		user.Admin = true
	}
	if adminRequest.Demote {
		user.Admin = false
	}
	updateUser(user)
	c.JSON(http.StatusAccepted, gin.H{"message": "updated all"})
}

func RefreshUserToken(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		c.AbortWithStatus(http.StatusNotAcceptable)
		return
	}
	authToken, err := refreshAuthToken(username.(string))
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.IndentedJSON(http.StatusOK, authToken)
}
