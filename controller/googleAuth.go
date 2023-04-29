package controller

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"github.com/Nivesh-Karma/go-user-admin/models"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
)

func GoogleLogin(c *gin.Context) {
	// this method validates the token from google
	// if token is valid then return the NK JWT token
	// Optionally add/update the user to DB

	//parse the token from header
	token := c.GetHeader("token")
	// get the id info uisng idtoken service from google
	idInfo, err := idtoken.Validate(c.Request.Context(), token, os.Getenv("CLIENT_ID"))
	// if error then return nil
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Unable to authenticate user.",
		})
		return
	}
	// get the username from email tag
	username := idInfo.Claims["email"].(string)
	// check if the user exists in DB
	userData, ok := findUser(username)
	// if user doesnt exist then create user using go routine
	if !ok {
		go func() {
			userRequest := models.UserRequest{}
			fName := idInfo.Claims["given_name"].(string)
			lName := idInfo.Claims["family_name"].(string)
			userRequest.FirstName = fName
			userRequest.LastName = lName
			userRequest.Username = username
			userRequest.SecurityQuestion1 = "Is the user logged from google?"
			userRequest.SecurityAnswer1 = "Yes"
			createUser(&userRequest, "google")
		}()
	} else {
		// if user exists but intially logged using email, convert them to google auth
		if userData.UserSource != "google" {
			go func() {
				userData.UserSource = "google"
				userData.Locked = false
				userData.Active = true
				userData.FailedCount = 0
				updateUser(userData)
			}()
		}
	}
	// generate the JWT token
	if token, err := createJWTToken(username); err != nil {
		log.Println("error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	} else {
		go createProfile(token.AccessToken)
		c.IndentedJSON(http.StatusOK, token)
	}
}

func createProfile(tokenString string) {
	url := os.Getenv("PROFILE_URL") + "/create-profile"
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	req, _ := http.NewRequest("POST", url, nil)
	client := &http.Client{Transport: tr}
	req.Header.Add("Authorization", "Bearer "+tokenString)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	if resp.StatusCode >= 400 {
		log.Println("Failed request")
	}
}
