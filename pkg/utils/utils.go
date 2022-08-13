package utils

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"github.com/whyxn/go-chat-backend/pkg/cache"
	"github.com/whyxn/go-chat-backend/pkg/core"
	"github.com/whyxn/go-chat-backend/pkg/helper"
	"github.com/whyxn/go-chat-backend/pkg/model"
	"github.com/whyxn/go-chat-backend/pkg/type"
	"strconv"
)

func GetAuthToken(c echo.Context) (string, error) {
	authToken := c.Request().Header.Get("Authorization")
	if len(authToken) == 0 {
		authToken = c.QueryParam("authToken")
	}

	if len(authToken) == 0 {
		return authToken, errors.New("Auth Token not found in the request")
	}

	return authToken, nil
}

func GetUserByAuthToken(authToken string) (*model.User, bool, error) {
	user := model.User{}
	isNewUser := true

	userIdStr, err := cache.GetCacheEngine().GetCache().Fetch(authToken)
	if err != nil {
		log.Warning("User not found in cache with auth token")

		userInfo := core.IsValidUser(authToken)
		if userInfo == nil || userInfo.Id == 0 {
			log.Info("Invalid auth token, no user found!")
			return nil, isNewUser, errors.New("Invalid auth token, no user found!")
		}

		UserInfoToUser(userInfo, &user)

		res, err := helper.FetchUserByUserId(user.UserId, true)
		if err != nil {

		} else if res.ID > 0 {
			user = *res
		}

	} else {

		userJsonStr, err := cache.GetCacheEngine().GetCache().Fetch("usr-" + userIdStr)
		if err != nil {
			log.Warning("User not found in cache with auth token")
		} else {

			err = json.Unmarshal([]byte(userJsonStr), &user)
			if err != nil {
				log.Error("Unmarshal user:", userJsonStr)
			}
		}

		if &user == nil || user.ID == 0 {

			userId, _ := strconv.ParseUint(userIdStr, 0, 64)

			res, err := helper.FetchUserByUserId(userId, true)

			if err != nil || &user == nil || user.ID == 0 {
				userInfo := core.IsValidUser(authToken)
				if userInfo == nil || userInfo.Id == 0 {
					log.Info("Invalid auth token, no user found!")
					return nil, isNewUser, errors.New("Invalid auth token, no user found!")
				}
				UserInfoToUser(userInfo, &user)
			} else if res.ID > 0 {
				user = *res
				isNewUser = false
			}

		} else {
			isNewUser = false
		}
	}

	return &user, isNewUser, nil
}

func UserInfoToUser(userInfo *core.User, user *model.User) *model.User {
	user.UserId = userInfo.Id
	user.Name = userInfo.Name
	user.Type = userInfo.Type
	user.ProfilePicture = userInfo.ProfilePicture
	user.Email = userInfo.Email
	user.PhoneNo = userInfo.PhoneNo

	return user
}

func GenerateSuccessResponse(data interface{}, message string, code string) _type.ResponseDTO {
	return _type.ResponseDTO{
		Status:  "success",
		Message: message,
		Code:    code,
		Data:    data,
	}
}

func GenerateErrorResponse(message string, code string) _type.ResponseDTO {
	return _type.ResponseDTO{
		Status:  "error",
		Message: message,
		Code:    code,
	}
}
