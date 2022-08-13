package db

import (
	"github.com/whyxn/go-chat-backend/pkg/model"
)

type DbOpsInterface interface {
	Save(obj interface{}) error
	GetUserByUserId(userId uint64, user model.User) (model.User, error)
	UpdateUserConnections(user *model.User) error
	GetChatWindowByParticipants(fromUser string, toUser string, chatWindow model.ChatWindow) (model.ChatWindow, error)
	GetChatWindowListByParticipant(participantUserId string) ([]model.ChatWindow, error)
	UpdateChatWindowLastMessageInfo(chatWindow *model.ChatWindow) error
}

type mysqlDbOps struct{}

var dbOpsInstance mysqlDbOps

func init() {
	dbOpsInstance = mysqlDbOps{}
}

func GetDbOps() DbOpsInterface {
	return &dbOpsInstance
}

func (md *mysqlDbOps) Save(obj interface{}) error {
	res := GetDB().Save(obj)
	return res.Error
}

func (md *mysqlDbOps) GetUserByUserId(userId uint64, user model.User) (model.User, error) {
	res := GetDB().Where("user_id = ?", userId).First(&user)
	return user, res.Error
}

func (md *mysqlDbOps) UpdateUserConnections(user *model.User) error {
	result := GetDB().Model(user).Updates(map[string]interface{}{"name": user.Name, "type": user.Type, "profile_picture": user.ProfilePicture, "email": user.Email, "phone_no": user.PhoneNo, "connections": user.Connections, "last_auth_token": user.LastAuthToken})
	return result.Error
}

func (md *mysqlDbOps) GetChatWindowByParticipants(fromUser string, toUser string, chatWindow model.ChatWindow) (model.ChatWindow, error) {
	res := GetDB().Preload("Participants").Where("participants_str IN ?", []string{fromUser + "&" + toUser, toUser + "&" + fromUser}).First(&chatWindow)
	return chatWindow, res.Error
}

func (md *mysqlDbOps) GetChatWindowListByParticipant(participantUserId string) ([]model.ChatWindow, error) {
	var chatWindowList []model.ChatWindow
	res := GetDB().Preload("Participants").Where("participants_str LIKE ?", "%&"+participantUserId).Or("participants_str LIKE ?", participantUserId+"&%").Order("last_message_at desc").Find(&chatWindowList)
	return chatWindowList, res.Error
}

func (md *mysqlDbOps) UpdateChatWindowLastMessageInfo(chatWindow *model.ChatWindow) error {
	res := GetDB().Model(chatWindow).Where("id = ?", chatWindow.ID).Updates(map[string]interface{}{"last_message_at": chatWindow.LastMessageAt, "last_message": chatWindow.LastMessage, "last_message_sender": chatWindow.LastMessageSender, "last_message_sender_user_id": chatWindow.LastMessageSenderUserId, "last_message_seen_by_recipient": chatWindow.LastMessageSeenByRecipient})
	return res.Error
}
