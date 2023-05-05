package controller

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
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
	token_type := c.GetHeader("token_type")
	username, fName, lName := "", "", ""
	if token_type == "One-tap" {
		username, fName, lName = OneTapLogin(c, token)
	} else if token_type == "Bearer" {
		username, fName, lName = VerifyGoogleToken(token)
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
		return
	}
	if username == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
		return
	}
	// check if the user exists in DB
	userData, ok := findUser(username)
	// if user doesnt exist then create user using go routine
	if !ok {
		go func() {
			userRequest := models.UserRequest{}
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
		resp := createProfile(token.AccessToken)
		if resp != nil {
			token.UserData.FirstName = fName
			token.UserData.LastName = lName
			token.UserData.Premium = userData.PremiumUser
			token.UserData.AddRatios = resp["add_ratios"]
		}
		c.IndentedJSON(http.StatusOK, token)
	}
}

func createProfile(tokenString string) map[string][]string {
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
		return nil
	}
	if resp.StatusCode >= 400 {
		log.Println("Failed request")
		return nil
	}
	defer resp.Body.Close()
	data := make(map[string][]string)
	json.NewDecoder(resp.Body).Decode(&data)
	return data
}

func GetProfile(tokenString string) models.ProfileResponse {
	url := os.Getenv("PROFILE_URL") + "/profile"
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	req, _ := http.NewRequest("GET", url, nil)
	client := &http.Client{Transport: tr}
	req.Header.Add("Authorization", "Bearer "+tokenString)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return models.ProfileResponse{}
	}
	if resp.StatusCode >= 400 {
		log.Println("Failed request")
		return models.ProfileResponse{}
	}
	defer resp.Body.Close()
	data := models.ProfileResponse{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Println(err)
		return models.ProfileResponse{}
	}
	return data
}

func OneTapLogin(c *gin.Context, token string) (string, string, string) {
	// get the id info uisng idtoken service from google
	idInfo, err := idtoken.Validate(c.Request.Context(), token, os.Getenv("CLIENT_ID"))
	// if error then return nil
	if err != nil {
		log.Println(err)
		return "", "", ""
	}
	// get the username from email tag
	username := idInfo.Claims["email"].(string)
	fName := idInfo.Claims["given_name"].(string)
	lName := idInfo.Claims["family_name"].(string)
	return username, fName, lName
}

func VerifyGoogleToken(token string) (string, string, string) {
	url := fmt.Sprintf("%s?access_token=%s", os.Getenv("GOOGLE_API"), token)
	req, _ := http.NewRequest("GET", url, nil)
	// ...
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	if resp.StatusCode >= 400 {
		return "", "", ""
	}
	defer resp.Body.Close()
	data := models.GoogleLoginRquest{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Println(err)
		return "", "", ""
	}
	return data.Email, data.GivenName, data.FamilyName
}
