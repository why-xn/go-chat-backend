package ws

import (
	"encoding/json"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/whyxn/go-chat-backend/pkg/cache"
	"github.com/whyxn/go-chat-backend/pkg/config"
	"github.com/whyxn/go-chat-backend/pkg/core"
	"github.com/whyxn/go-chat-backend/pkg/helper"
	"github.com/whyxn/go-chat-backend/pkg/model"
	"github.com/whyxn/go-chat-backend/pkg/utils"
	"net/http"
	"strconv"
	"syscall"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func autoId() string {
	var err error
	return uuid.Must(uuid.NewV4(), err).String()
}

var epoller *epoll

func InitEpoller() {
	// Start epoll
	var err error
	epoller, err = MkEpoll()
	if err != nil {
		panic(err)
	}
}

func WsHandler(w http.ResponseWriter, r *http.Request) {

	keys, ok := r.URL.Query()["authToken"]

	if !ok || len(keys[0]) < 1 {
		log.Warning("Url Param 'authToken' is missing")
		return
	}

	// Query()["key"] will return an array of items,
	// we only want the single item.

	authToken := keys[0]

	if &authToken != nil && len(authToken) > 0 {

		user := model.User{}
		isNewUser := true

		userIdStr, err := cache.GetCacheEngine().GetCache().Fetch(authToken)
		if err != nil {
			log.Warning("User not found in cache with auth token")

			userInfo := core.IsValidUser(authToken)
			if userInfo == nil || userInfo.Id == 0 {
				log.Warning("Invalid auth token, no user found!")
				return
			}

			utils.UserInfoToUser(userInfo, &user)

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
						log.Warning("Invalid auth token, no user found!")
						return
					}
					utils.UserInfoToUser(userInfo, &user)
				} else if res.ID > 0 && res.UserId > 0 {
					user = *res
					isNewUser = false
				} else {
					userInfo := core.IsValidUser(authToken)
					if userInfo == nil || userInfo.Id == 0 {
						log.Warning("Invalid auth token, no user found!")
						return
					}
					utils.UserInfoToUser(userInfo, &user)
				}

			} else {
				isNewUser = false
			}
		}

		if user.UserId == 0 {
			log.Error("Zero userid found!")
			return
		}

		user.LastAuthToken = authToken

		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true

		}

		// Upgrade connection
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Println(err)
			return
		}

		client := core.Client{
			Connection: conn,
			Id:         core.AutoId(),
			UserId:     user.ID,
		}

		if err := epoller.Add(client); err != nil {
			log.Error("Failed to add connection %v", err)
			conn.Close()
		} else {
			log.Info("New client added to epoller")
		}

		// add this client into the list
		core.GetChatEngine().GetGoChat().AddClient(&user, client, isNewUser, authToken)

		log.Info("New Client is connected, total: ", len(core.GetChatEngine().GetGoChat().ClientsMap))

	} else {
		log.Error("Empty auth token!")
	}
}

func StartListeningFromClientConnections() {
	for {
		clients, err := epoller.Wait()
		if err != nil {
			log.Debugln("Failed to epoll wait %v", err)
			continue
		}

		if len(clients) > 0 {
			log.Debugln("New message might be available")
		}
		for _, client := range clients {
			if client.Connection == nil {
				break
			}
			if msg, _, err := wsutil.ReadClientData(client.Connection); err != nil {
				if err := epoller.Remove(client); err != nil {
					log.Error("Failed to remove %v", err)
				}
				log.Debugln("Removing client")
				client.Connection.Close()
			} else {
				log.Debugln("NEW MSG:", string(msg))
				core.GetChatEngine().GetGoChat().HandleReceiveMessage(client, 1, msg)
			}
		}
	}
}

func StartWebsocketServer() {
	// Increase resources limitations
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	InitEpoller()
	go StartListeningFromClientConnections()

	go core.PingAllClients()

	http.HandleFunc("/", WsHandler)
	if err := http.ListenAndServe("0.0.0.0:"+config.GetWsServerPort(), nil); err != nil {
		log.Fatal(err)
	}
}
