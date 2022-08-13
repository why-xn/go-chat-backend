package core

import (
	"encoding/json"
	"errors"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/whyxn/go-chat-backend/pkg/cache"
	"github.com/whyxn/go-chat-backend/pkg/config"
	"github.com/whyxn/go-chat-backend/pkg/db"
	"github.com/whyxn/go-chat-backend/pkg/helper"
	"github.com/whyxn/go-chat-backend/pkg/model"
	"github.com/whyxn/go-chat-backend/pkg/pubsub"
	"github.com/whyxn/go-chat-backend/pkg/type"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	INIT       = "init"
	SEND       = "send"
	DISCONNECT = "disconnect"
)

type ChatEngineInterface interface {
	init()
	GetGoChat() *GoChat
}

type goChatEngine struct {
	GoChat GoChat
}

// Implementing Singleton
var singletonChatEngine *goChatEngine
var onceChatEngine sync.Once

func GetChatEngine() *goChatEngine {
	onceChatEngine.Do(func() {
		log.Info("Initializing Singleton GoChat Engine")
		singletonChatEngine = &goChatEngine{}
		singletonChatEngine.init()
	})
	return singletonChatEngine
}

func (cm *goChatEngine) init() {
	cm.GoChat = GoChat{}
	cm.GoChat.ClientsMap = map[string]Client{}
	log.Info("Initialized GoChat Engine")
}

func (cm *goChatEngine) GetGoChat() *GoChat {
	return &cm.GoChat
}

type GoChatInterface interface {
	AddClient(user *model.User, client Client, isNewUser bool, authToken string) *GoChat
	RemoveClient(client Client) *GoChat
	SendMessage(fromClientId string, chatWindowId string, message string) error
	InitChatWindow(fromClient *Client, toUserId uint64) (string, error)
	InitChatWindowByApi(fromUser *model.User, toUserId uint64) (string, error)
	HandleReceiveMessage(client Client, messageType int, payload []byte) *GoChat
}

type GoChat struct {
	ClientsMap map[string]Client
}

type ClientInterface interface {
	Send(message []byte) error
}

type Client struct {
	Id         string
	Connection net.Conn
	UserId     uint64
}

func (cl *GoChat) AddClient(user *model.User, client Client, isNewUser bool, authToken string) *GoChat {

	cl.ClientsMap[client.Id] = client

	log.Info("adding new client to the list", client.Id, len(cl.ClientsMap))

	helper.AddConnection(client.Id, user.UserId)

	user.Connections = user.Connections + client.Id + ";"

	if isNewUser {
		err := db.GetDbOps().Save(user)
		if err != nil {
			log.Error("While saving new user data in DB", err.Error())
			return cl
		}
	} else {
		err = db.GetDbOps().UpdateUserConnections(user)
		if err != nil {
			log.Error("While updating user connections in DB", err.Error())
			return cl
		}
	}

	userIdStr := strconv.FormatUint(user.UserId, 10)

	err = cache.GetCacheEngine().GetCache().Save("conu-"+client.Id, userIdStr, 30*time.Minute)
	if err != nil {
		log.Error("While saving user data in cache", err)
	}

	newUserByte, err := json.Marshal(user)
	if err != nil {
		log.Error("While json marshal user data", err.Error())
	} else {
		_ = cache.GetCacheEngine().GetCache().Save(authToken, userIdStr, 30*time.Minute)
		_ = cache.GetCacheEngine().GetCache().Save("usr-"+userIdStr, string(newUserByte), 30*time.Minute)
	}

	userId := strconv.FormatUint(user.UserId, 10)

	out := _type.OutGoingMessage{
		Event: "WELCOME",
		Payload: map[string]string{
			"userId":       userId,
			"connectionId": client.Id,
		},
	}

	payload, err := json.Marshal(out)

	//payload := []byte("Hello, user id: " + userId + ". Your connection id: " + client.Id)

	client.Send(payload)

	return cl
}

func (cl *GoChat) RemoveClient(client Client) *GoChat {

	log.Info("Removing Client:", client.Id)
	// first remove all subscriptions by this client

	if _, ok := cl.ClientsMap[client.Id]; ok {
		//do something here
		delete(cl.ClientsMap, client.Id)
		helper.PermanentlyDeleteConnection(client.Id)
		cache.GetCacheEngine().GetCache().Delete("conu-" + client.Id)

		user, err := helper.FetchUserByConnectionId(client.Id, true)
		if err != nil {
			log.Error("While fetching user by connection id", client.Id, err.Error())
		} else {
			userIdStr := strconv.FormatUint(user.UserId, 10)
			if len(user.Connections) > 0 && strings.Contains(user.Connections, client.Id) {
				user.Connections = strings.ReplaceAll(user.Connections, client.Id+";", "")

				err = db.GetDbOps().UpdateUserConnections(user)
				if err != nil {
					log.Error("While updating user connections in DB", err.Error())
				}

				newUserByte, err := json.Marshal(user)
				if err != nil {
					log.Error("While json marshal user data", err.Error())
				} else {
					_ = cache.GetCacheEngine().GetCache().Save("usr-"+userIdStr, string(newUserByte), 30*time.Minute)
				}
			}
		}
	}

	return cl
}

func (cl *GoChat) SendMessage(fromClientId string, chatWindowId string, message string) error {

	chatWindow, err := helper.FetchChatWindow(chatWindowId, false)
	if err != nil {
		return err
	}

	fromUser, err := helper.FetchUserByConnectionId(fromClientId, false)
	if err != nil {
		return err
	}

	requesterHasPermission := false

	for _, participant := range chatWindow.Participants {
		if participant.UserId == fromUser.UserId {
			requesterHasPermission = true
			break
		}
	}

	if !requesterHasPermission {
		return errors.New("Permission denied")
	}

	chatMessage := model.ChatMessage{
		Type:            "txt",
		Message:         message,
		Sender:          fromUser.UserId,
		ChatWindowRefer: chatWindow.Uid,
	}

	err = db.GetDbOps().Save(&chatMessage)
	if err != nil {
		log.Error("Failed to save message in db", err.Error())
		return err
	}

	now := time.Now()

	chatWindow.LastMessageAt = &now
	chatWindow.LastMessage = message
	chatWindow.LastMessageSender = fromUser.Name
	chatWindow.LastMessageSenderUserId = fromUser.UserId
	chatWindow.LastMessageSeenByRecipient = false

	err = db.GetDbOps().UpdateChatWindowLastMessageInfo(chatWindow)
	if err != nil {
		log.Error("Failed to update chat window", err.Error())
	}

	outGoingMessage := _type.OutGoingMessage{
		Event:        "NEW_MESSAGE",
		ChatWindowId: chatWindow.Uid,
		From:         fromUser.UserId,
		Payload: _type.MessagePayload{
			Type:    "txt",
			Message: message,
		},
	}

	payloadByte, err := json.Marshal(outGoingMessage)

	var toUserId uint64

	for _, participant := range chatWindow.Participants {
		if participant.UserId != fromUser.UserId {
			toUserId = participant.UserId
		}
	}

	toUser, err := helper.FetchUserByUserId(toUserId, false)
	if err != nil {
		return err
	}

	fromConnections := helper.GetUserConnectionsAsArray(fromUser)
	toConnections := helper.GetUserConnectionsAsArray(toUser)

	for _, connection := range fromConnections {
		if fromClientId != connection {
			if _, ok := cl.ClientsMap[connection]; ok {
				client := cl.ClientsMap[connection]
				client.Send(payloadByte)
			} else {
				outGoingMessage.ToConnection = connection
				pub := pubsub.Service.Publish(config.GetRedisMessageOutgoingChannel(), outGoingMessage)
				if err = pub.Err(); err != nil {
					log.Error("PublishString() error", err)
				}
			}
		}
	}

	for _, connection := range toConnections {
		if _, ok := cl.ClientsMap[connection]; ok {
			client := cl.ClientsMap[connection]
			client.Send(payloadByte)
		} else {
			outGoingMessage.ToConnection = connection
			pub := pubsub.Service.Publish(config.GetRedisMessageOutgoingChannel(), outGoingMessage)
			if err = pub.Err(); err != nil {
				log.Error("PublishString() error", err)
			}
		}
	}

	return nil
}

func AutoId() string {
	var err error
	return uuid.Must(uuid.NewV4(), err).String()
}

func preValidateInitChatWindowRequest(fromUser *model.User, toUser *model.User) error {
	if fromUser.Type == toUser.Type {
		return errors.New("chat cannot be established between similar types of users")
	}
	return nil
}

func (cl *GoChat) InitChatWindowByApi(fromUser *model.User, toUserId uint64) (string, error) {

	toUser, err := helper.FetchUserByUserId(toUserId, true)
	if err != nil {
		return "", err
	} else if toUser.UserId == 0 {
		log.Error("To user not found to init chat window:", toUserId)
		return "", errors.New("To user not found to init chat window")
	}

	err = preValidateInitChatWindowRequest(fromUser, toUser)
	if err != nil {
		return "", err
	}

	fromUserIdStr := strconv.FormatUint(fromUser.UserId, 10)
	toUserIdStr := strconv.FormatUint(toUser.UserId, 10)

	var chatWindow model.ChatWindow

	chatWindow, err = db.GetDbOps().GetChatWindowByParticipants(fromUserIdStr, toUserIdStr, chatWindow)
	if err != nil || chatWindow.ID == 0 {
		log.Warnf("Chat Window doesn't exists for %v & %v. %v", fromUserIdStr, toUserIdStr, err)

		if fromUser.Type == "exhibitor" {
			return "", errors.New("An exhibitor cannot initiate a chat with a visitor.")
		}

		log.Infof("Creating new chat window for %v & %v", fromUserIdStr, toUserIdStr)

		chatWindow = model.ChatWindow{
			Uid: AutoId(),
			Participants: []model.User{
				*fromUser,
				*toUser,
			},
			ParticipantsStr: fromUserIdStr + "&" + toUserIdStr,
			LastMessageAt:   nil,
		}

		err = db.GetDbOps().Save(&chatWindow)
		if err != nil && chatWindow.ID == 0 {
			return "", err
		}
	}

	chatWindowJsonByte, err := json.Marshal(chatWindow)
	if err != nil {
		return "", err
	}

	err = cache.GetCacheEngine().GetCache().Save("cw-"+chatWindow.Uid, string(chatWindowJsonByte), 10*time.Minute)
	if err != nil {
		log.Error(err.Error())
	}

	outGoingMessage := _type.OutGoingMessage{
		Event:        "INIT_CHAT",
		ChatWindowId: chatWindow.Uid,
		From:         0,
		Payload: map[string]interface{}{
			"participants": chatWindow.Participants,
		},
	}

	payloadByte, err := json.Marshal(outGoingMessage)

	fromConnections := helper.GetUserConnectionsAsArray(fromUser)
	toConnections := helper.GetUserConnectionsAsArray(toUser)

	for _, connection := range fromConnections {
		if _, ok := cl.ClientsMap[connection]; ok {
			client := cl.ClientsMap[connection]
			client.Send(payloadByte)
		} else {
			outGoingMessage.ToConnection = connection
			pub := pubsub.Service.Publish(config.GetRedisMessageOutgoingChannel(), outGoingMessage)
			if err = pub.Err(); err != nil {
				log.Error("PublishString() error", err)
			}
		}
	}

	for _, connection := range toConnections {
		if _, ok := cl.ClientsMap[connection]; ok {
			client := cl.ClientsMap[connection]
			client.Send(payloadByte)
		} else {
			outGoingMessage.ToConnection = connection
			pub := pubsub.Service.Publish(config.GetRedisMessageOutgoingChannel(), outGoingMessage)
			if err = pub.Err(); err != nil {
				log.Error("PublishString() error", err)
			}
		}
	}

	return chatWindow.Uid, nil
}

func (cl *GoChat) InitChatWindow(fromClient *Client, toUserId uint64) (string, error) {

	fromUser, err := helper.FetchUserByConnectionId(fromClient.Id, true)
	if err != nil {
		return "", err
	}

	return cl.InitChatWindowByApi(fromUser, toUserId)
}

func (cl *GoChat) HandleReceiveMessage(client Client, messageType int, payload []byte) *GoChat {

	m := _type.Message{}

	err := json.Unmarshal(payload, &m)
	if err != nil {
		log.Error("This is not a correct message payload.", string(payload))
		return cl
	}

	log.Info(m.Action)

	switch m.Action {
	case INIT:
		log.Info("Client wants to chat with", m.ToUser)

		chatWindowId, err := cl.InitChatWindow(&client, m.ToUser)
		if err != nil {
			log.Error("While Init chat window:", err.Error())
		}

		if len(chatWindowId) > 0 {
			log.Infof("Chat window initialized: %v", chatWindowId)
		}

		break

	case SEND:
		cl.SendMessage(client.Id, m.ChatWindowId, m.Message)
		break

	case DISCONNECT:
		cl.RemoveClient(client)
		break

	default:
		break
	}

	return cl
}

func (client *Client) Send(message []byte) error {
	return wsutil.WriteServerMessage(client.Connection, ws.OpText, message)
}
