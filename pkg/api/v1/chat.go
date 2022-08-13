package v1

import (
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"github.com/whyxn/go-chat-backend/pkg/core"
	"github.com/whyxn/go-chat-backend/pkg/db"
	"github.com/whyxn/go-chat-backend/pkg/helper"
	"github.com/whyxn/go-chat-backend/pkg/model"
	"github.com/whyxn/go-chat-backend/pkg/type"
	"github.com/whyxn/go-chat-backend/pkg/utils"
	"net/http"
	"strconv"
)

type ChatApiInterface interface {
	InitChatWindow(c echo.Context) error
	SendMessage(c echo.Context) error
	GetMessages(c echo.Context) error
	GetAllUsers(c echo.Context) error
	GetMe(c echo.Context) error
	GetChatWindow(c echo.Context) error
	GetChatWindowList(c echo.Context) error
}

type chatApi struct{}

func NewChatApi() ChatApiInterface {
	return &chatApi{}
}

func (ch *chatApi) GetMe(c echo.Context) error {
	authToken, err := utils.GetAuthToken(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	requester, _, err := utils.GetUserByAuthToken(authToken)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	return c.JSON(http.StatusOK, utils.GenerateSuccessResponse(requester, "", ""))

}

func (ch *chatApi) GetAllUsers(c echo.Context) error {
	authToken, err := utils.GetAuthToken(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	requester, _, err := utils.GetUserByAuthToken(authToken)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	if requester.Type == "exhibitor" {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse("Permission Denied", ""))
	}

	users := []model.User{}

	res := db.GetDB().Where("user_id != ? AND type != ?", requester.UserId, requester.Type).Find(&users)
	if res.Error != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(res.Error.Error(), ""))
	}

	return c.JSON(http.StatusOK, utils.GenerateSuccessResponse(users, "", ""))

}

func (ch *chatApi) GetChatWindow(c echo.Context) error {
	authToken, err := utils.GetAuthToken(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	requester, _, err := utils.GetUserByAuthToken(authToken)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	chatWindowUid := c.Param("id")

	chatWindow, err := helper.FetchChatWindow(chatWindowUid, true)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	requesterHasPermission := false

	for _, participant := range chatWindow.Participants {
		if participant.UserId == requester.UserId {
			requesterHasPermission = true
			break
		}
	}

	if !requesterHasPermission {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse("Permission denied", ""))
	}

	return c.JSON(http.StatusOK, utils.GenerateSuccessResponse(chatWindow, "", ""))
}

func (ch *chatApi) GetChatWindowList(c echo.Context) error {
	authToken, err := utils.GetAuthToken(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	requester, _, err := utils.GetUserByAuthToken(authToken)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	requesterUserIdStr := strconv.FormatUint(requester.UserId, 10)

	chatWindowList, err := db.GetDbOps().GetChatWindowListByParticipant(requesterUserIdStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.GenerateErrorResponse(err.Error(), ""))
	}

	return c.JSON(http.StatusOK, utils.GenerateSuccessResponse(chatWindowList, "", ""))
}

func (ch *chatApi) InitChatWindow(c echo.Context) error {
	authToken, err := utils.GetAuthToken(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	requester, _, err := utils.GetUserByAuthToken(authToken)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	var input _type.InitChatWindowInput
	if err := c.Bind(&input); err != nil {
		log.Error("Input Error:", err.Error())
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	chatWindowUid, err := core.GetChatEngine().GetGoChat().InitChatWindowByApi(requester, input.ToUser)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	return c.JSON(http.StatusOK, utils.GenerateSuccessResponse(map[string]string{"chatWindowId": chatWindowUid}, "", ""))
}

func (ch *chatApi) SendMessage(c echo.Context) error {
	authToken, err := utils.GetAuthToken(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	_, _, err = utils.GetUserByAuthToken(authToken)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	chatWindowUid := c.Param("id")

	var input _type.SendMessageInput
	if err := c.Bind(&input); err != nil {
		log.Error("Input Error:", err.Error())
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	err = core.GetChatEngine().GetGoChat().SendMessage(input.ConnectionId, chatWindowUid, input.Message)
	if err != nil {
		log.Error("Error:", err.Error())
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	return c.JSON(http.StatusOK, utils.GenerateSuccessResponse(nil, "Sent", ""))
}

func (ch *chatApi) GetMessages(c echo.Context) error {
	authToken, err := utils.GetAuthToken(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	requester, _, err := utils.GetUserByAuthToken(authToken)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	chatWindowUid := c.Param("id")

	chatWindow, err := helper.FetchChatWindow(chatWindowUid, true)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(err.Error(), ""))
	}

	requesterHasPermission := false

	for _, participant := range chatWindow.Participants {
		if participant.UserId == requester.UserId {
			requesterHasPermission = true
			break
		}
	}

	if !requesterHasPermission {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse("Permission denied", ""))
	}

	chatMessages := []model.ChatMessage{}

	res := db.GetDB().Where("chat_window_refer = ?", chatWindowUid).Find(&chatMessages)
	if res.Error != nil {
		return c.JSON(http.StatusBadRequest, utils.GenerateErrorResponse(res.Error.Error(), ""))
	}

	if requester.UserId != chatWindow.LastMessageSenderUserId {
		res = db.GetDB().Model(chatWindow).Update("last_message_seen_by_recipient", true)
		if res.Error != nil {
			log.Error(res.Error.Error())
		}
	}

	return c.JSON(http.StatusOK, utils.GenerateSuccessResponse(chatMessages, "", ""))
}
