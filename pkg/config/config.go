package config

import (
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
)

var runMode string
var httpServerPort string
var wsServerPort string
var redisHost string
var redisPort string
var redisPassword string
var mysqlHost string
var mysqlPort string
var mysqlDb string
var mysqlUsername string
var mysqlPassword string
var authTokenValidationEndPoint string
var redisMessageOutgoingChannel string

func LoadEnvironmentVariables() {
	log.Info("Loading environment variables")

	runMode = os.Getenv("RUN_MODE")
	if runMode == "" {
		runMode = "develop"
	}

	log.Info("RUN MODE:", runMode)

	if runMode != "production" {
		//Load .env file
		err := godotenv.Load()
		if err != nil {
			log.Error("Loading environment file:", err.Error())
			os.Exit(1)
		}
	}

	httpServerPort = os.Getenv("HTTP_SERVER_PORT")
	wsServerPort = os.Getenv("WS_SERVER_PORT")
	redisHost = os.Getenv("REDIS_HOST")
	redisPort = os.Getenv("REDIS_PORT")
	redisPassword = os.Getenv("REDIS_PASSWORD")
	mysqlHost = os.Getenv("MYSQL_HOST")
	mysqlPort = os.Getenv("MYSQL_PORT")
	mysqlDb = os.Getenv("MYSQL_DB")
	mysqlUsername = os.Getenv("MYSQL_USERNAME")
	mysqlPassword = os.Getenv("MYSQL_PASSWORD")
	authTokenValidationEndPoint = os.Getenv("AUTH_TOKEN_VALIDATION_ENDPOINT")
	redisMessageOutgoingChannel = os.Getenv("REDIS_MESSAGE_OUTGOING_CHANNEL")
}

func GetHttpServerPort() string {
	return httpServerPort
}

func GetWsServerPort() string {
	return wsServerPort
}

func GetRedisHost() string {
	return redisHost
}

func GetRedisPort() string {
	return redisPort
}

func GetRedisPassword() string {
	return redisPassword
}

func GetMySqlHost() string {
	return mysqlHost
}

func GetMySqlPort() string {
	return mysqlPort
}

func GetMySqlDB() string {
	return mysqlDb
}

func GetMySqlUsername() string {
	return mysqlUsername
}

func GetMySqlPassword() string {
	return mysqlPassword
}

func GetAuthTokenValidationEndPoint() string {
	return authTokenValidationEndPoint
}

func GetRedisMessageOutgoingChannel() string {
	return redisMessageOutgoingChannel
}
