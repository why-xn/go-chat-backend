package helper

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/whyxn/go-chat-backend/pkg/cache"
	"github.com/whyxn/go-chat-backend/pkg/db"
	"github.com/whyxn/go-chat-backend/pkg/model"
	"time"
)

func FetchChatWindow(chatWindowId string, ignoreCache bool) (*model.ChatWindow, error) {
	var chatWindow model.ChatWindow
	needToFetchFromDb := false

	if ignoreCache {
		needToFetchFromDb = true
	} else {
		chatWindowJsonStr, err := cache.GetCacheEngine().GetCache().Fetch("cw-" + chatWindowId)
		if err != nil {
			log.Error("Chat window not found in cache", chatWindowId)
			needToFetchFromDb = true

		} else {
			err = json.Unmarshal([]byte(chatWindowJsonStr), &chatWindow)
			if err != nil {
				log.Error("Unmarshal chat window", chatWindowId)
				needToFetchFromDb = true
			}
			if chatWindow.Participants == nil || len(chatWindow.Participants) == 0 {
				needToFetchFromDb = true
			}
		}
	}

	if needToFetchFromDb {
		result := db.GetDB().Preload("Participants").Where("uid = ?", chatWindowId).First(&chatWindow)
		if result.Error != nil {
			log.Error("Chat window doesn't exists", result.Error.Error())
			return nil, errors.New("Chat window doesn't exists")
		}

		if &chatWindow == nil || chatWindow.ID == 0 {
			log.Error("Chat window doesn't exists")
			return nil, errors.New("Chat window doesn't exists")
		}

		chatWindowJsonByte, err := json.Marshal(chatWindow)
		if err == nil {
			err = cache.GetCacheEngine().GetCache().Save("cw-"+chatWindow.Uid, string(chatWindowJsonByte), 10*time.Minute)
			if err != nil {
				log.Error("Failed saving chat window in cache")
			}
		}
	}

	return &chatWindow, nil
}
