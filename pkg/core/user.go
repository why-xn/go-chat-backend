package core

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/whyxn/go-chat-backend/pkg/config"
	"io/ioutil"
	"net/http"
)

type User struct {
	Id             uint64
	Name           string
	Type           string
	Password       string
	AuthToken      string
	ProfilePicture string `json:"profile_picture"`
	Email          string
	PhoneNo        string
}

type UserDataResponse struct {
	StatusCode int64  `json:"statusCode"`
	Message    string `json:"message"`
	Data       User   `json:"data"`
}

var users map[string]User

func InitDummyUsers() {

	users = map[string]User{}

	du1 := User{
		Id:        1,
		Name:      "visitor1",
		Password:  "hello123",
		AuthToken: "visitor1",
		Email:     "visitor1@chat.com",
		PhoneNo:   "0123456789",
		Type:      "visitor",
	}

	users[du1.AuthToken] = du1

	du2 := User{
		Id:        2,
		Name:      "visitor2",
		Password:  "hello123",
		AuthToken: "visitor2",
		Email:     "visitor2@dchat.com",
		PhoneNo:   "0123456789",
		Type:      "visitor",
	}

	users[du2.AuthToken] = du2

	du3 := User{
		Id:        3,
		Name:      "exhibitor1",
		Password:  "hello123",
		AuthToken: "exhibitor1",
		Email:     "exhibitor1@chat.com",
		PhoneNo:   "0123456789",
		Type:      "exhibitor",
	}

	users[du3.AuthToken] = du3

	du4 := User{
		Id:        4,
		Name:      "exhibitor2",
		Password:  "hello123",
		AuthToken: "exhibitor2",
		Email:     "exhibitor2@chat.com",
		PhoneNo:   "0123456789",
		Type:      "exhibitor",
	}

	users[du4.AuthToken] = du4

}

func IsValidUser(autToken string) *User {
	// Allowing the dummy users to pass for testing purpose
	if _, ok := users[autToken]; ok {
		User := users[autToken]
		return &User
	}

	url := config.GetAuthTokenValidationEndPoint()

	// Create a Bearer string by appending string access token
	var bearer = "Bearer " + autToken

	// Create a new request using http
	req, err := http.NewRequest("POST", url, nil)

	// add authorization header to the req
	req.Header.Add("Authorization", bearer)

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("On response while fetching user data -", err)
		return nil
	}

	body, _ := ioutil.ReadAll(resp.Body)
	log.Debugln(string([]byte(body)))

	var userDataResponse UserDataResponse

	err = json.Unmarshal([]byte(body), &userDataResponse)
	if err != nil {
		log.Error("While unmarshalling user data response", err.Error())
		return nil
	}

	if userDataResponse.StatusCode == 800200 {
		return &userDataResponse.Data
	}

	return nil

}
