package db

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whyxn/go-chat-backend/pkg/config"
	"github.com/whyxn/go-chat-backend/pkg/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var err error

func InitDbConnection() {
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True", config.GetMySqlUsername(), config.GetMySqlPassword(), config.GetMySqlHost(), config.GetMySqlPort(), config.GetMySqlDB())
	dsn = dsn + "&loc=Asia%2FDhaka"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatal("Failed to connect to mysql", err.Error())
	} else {
		log.Info("MySql connection established")
	}

	db.AutoMigrate(&model.User{})
	db.AutoMigrate(&model.Connection{})
	db.AutoMigrate(&model.ChatWindow{})
	db.AutoMigrate(&model.ChatMessage{})
	db.AutoMigrate(&model.ChatSeenStatus{})
}

func GetDB() *gorm.DB {
	return db
}
