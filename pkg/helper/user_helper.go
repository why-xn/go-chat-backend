package helper

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/whyxn/go-chat-backend/pkg/cache"
	"github.com/whyxn/go-chat-backend/pkg/db"
	"github.com/whyxn/go-chat-backend/pkg/model"
	"strconv"
	"strings"
)

func FetchUserById(userId uint64, ignoreCache bool) (*model.User, error) {

	var user model.User
	needToFetchFromDb := false
	var userIdStr string

	userIdStr = strconv.FormatUint(userId, 10)

	if ignoreCache {
		needToFetchFromDb = true
	} else {
		userJsonStr, err := cache.GetCacheEngine().GetCache().Fetch("usr-" + userIdStr)
		if err != nil {
			log.Error("User not found in cache", userIdStr)
			needToFetchFromDb = true

		} else {
			err = json.Unmarshal([]byte(userJsonStr), &user)
			if err != nil {
				log.Error("Unmarshal user", userIdStr)
				needToFetchFromDb = true
			}
		}
	}

	if needToFetchFromDb {
		result := db.GetDB().Find(&user, userId)
		if result.Error != nil {
			log.Error("While finding user in DB", result.Error.Error())
			return nil, errors.New(result.Error.Error())
		}
	}

	return &user, nil

}

func FetchUserByUserId(userId uint64, ignoreCache bool) (*model.User, error) {

	var user model.User
	needToFetchFromDb := false
	var userIdStr string
	userIdStr = strconv.FormatUint(userId, 10)

	if ignoreCache {
		needToFetchFromDb = true
	} else {
		userJsonStr, err := cache.GetCacheEngine().GetCache().Fetch("usr-" + userIdStr)
		if err != nil {
			log.Warning("User not found in cache", userIdStr)
			needToFetchFromDb = true

		} else {
			err = json.Unmarshal([]byte(userJsonStr), &user)
			if err != nil {
				log.Warning("Unmarshal user", userIdStr)
				needToFetchFromDb = true
			}
		}
	}

	if needToFetchFromDb {
		result := db.GetDB().Where("user_id = ?", userId).Find(&user)
		if result.Error != nil {
			log.Error("User not found -", result.Error.Error())
			return nil, errors.New(result.Error.Error())
		}
	}

	return &user, nil

}

func FetchUserByConnectionId(connectionId string, ignoreCache bool) (*model.User, error) {

	var user model.User
	needToFetchFromDb := false
	var userIdStr string
	var err error

	if ignoreCache {
		needToFetchFromDb = true
	} else {
		userIdStr, err = cache.GetCacheEngine().GetCache().Fetch("conu-" + connectionId)
		if err != nil {
			log.Error("User id not found in cache", userIdStr)
			needToFetchFromDb = true
		} else {
			userJsonStr, err := cache.GetCacheEngine().GetCache().Fetch("usr-" + userIdStr)
			if err != nil {
				log.Warning("User not found in cache", userIdStr)
				needToFetchFromDb = true
			} else {
				err = json.Unmarshal([]byte(userJsonStr), &user)
				if err != nil {
					log.Warning("Unmarshal user", userIdStr)
					needToFetchFromDb = true
				}
			}
		}
	}

	if needToFetchFromDb {
		result := db.GetDB().Where("connections LIKE ?", "%"+connectionId+"%").First(&user)
		if result.Error != nil {
			log.Error("User not found", result.Error.Error())
			return nil, result.Error
		}

		if &user == nil || user.ID == 0 {
			log.Error("User not found")
			return nil, errors.New("user not found")
		}
	}

	return &user, nil

}

func AddConnection(connectionUid string, userId uint64) {
	connection := model.Connection{
		Uid:       connectionUid,
		UserRefer: userId,
	}

	res := db.GetDB().Save(&connection)
	if res.Error != nil {
		log.Println(res.Error.Error())
	}

}

func PermanentlyDeleteConnection(connectionUid string) {
	res := db.GetDB().Unscoped().Where("uid = ?", connectionUid).Delete(&model.Connection{})
	if res.Error != nil {
		log.Println(res.Error.Error())
	}
}

func SoftDeleteConnection(connectionUid string) {
	res := db.GetDB().Where("uid = ?", connectionUid).Delete(&model.Connection{})
	if res.Error != nil {
		log.Println(res.Error.Error())
	}
}

func GetUserConnectionsAsArray(user *model.User) []string {
	var connections []string
	if len(user.Connections) > 0 {
		connections = strings.Split(user.Connections, ";")
		if len(connections) > 1 {
			connections = connections[:len(connections)-1]
		}
	}
	return connections
}
