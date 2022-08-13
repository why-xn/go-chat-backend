package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/whyxn/go-chat-backend/pkg/cache"
	"github.com/whyxn/go-chat-backend/pkg/config"
	"github.com/whyxn/go-chat-backend/pkg/core"
	"github.com/whyxn/go-chat-backend/pkg/db"
	"github.com/whyxn/go-chat-backend/pkg/handler"
	"github.com/whyxn/go-chat-backend/pkg/pubsub"
	"github.com/whyxn/go-chat-backend/pkg/router"
	"github.com/whyxn/go-chat-backend/pkg/server"
	"github.com/whyxn/go-chat-backend/pkg/ws"
	_ "net/http/pprof"
	"time"
)

func configureLogrus() {
	log.SetLevel(log.InfoLevel)
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
}

func main() {

	timezoneLocString := "Asia/Dhaka"
	time.LoadLocation(timezoneLocString)

	// configure log formatting
	configureLogrus()

	config.LoadEnvironmentVariables()

	// initialize DB connection
	db.InitDbConnection()

	// initialize managers
	cache.GetCacheEngine()
	cache.GetCacheEngine().GetCache().Flush()
	core.GetChatEngine()

	// initialize few dummy users to test purpose
	core.InitDummyUsers()

	// Subscribe to Outgoing Message Channel
	// (This is for PubSub between the server instances while running multiple server instances)
	_, err := pubsub.NewSubscriber(config.GetRedisMessageOutgoingChannel(), handler.OutGoingHandler)
	if err != nil {
		log.Fatalln("Failed to Subscribe to Outgoing Message Channel!", err)
	}

	// Start WebSocket Server (Epoll) in another Thread
	go ws.StartWebsocketServer()

	// Start Http Server
	srv := server.New()
	router.Routes(srv)
	srv.Logger.Fatal(srv.Start(":" + config.GetHttpServerPort()))

}
