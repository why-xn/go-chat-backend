package cache

import (
	"github.com/faabiosr/cachego"
	"github.com/faabiosr/cachego/redis"
	log "github.com/sirupsen/logrus"
	"github.com/whyxn/go-chat-backend/pkg/config"
	rd "gopkg.in/redis.v4"
	"sync"
)

type CacheEngineInterface interface {
	initConnection()
	GetCache() cachego.Cache
}

type cacheEngine struct {
	cache cachego.Cache
}

// Implementing Singleton
var singletonCacheEngine *cacheEngine
var onceCacheEngine sync.Once

func GetCacheEngine() *cacheEngine {
	onceCacheEngine.Do(func() {
		log.Info("Starting Initializing Singleton Cache Engine...")
		singletonCacheEngine = &cacheEngine{}
		singletonCacheEngine.initConnection()
	})
	return singletonCacheEngine
}

func (cm *cacheEngine) initConnection() {
	cm.cache = redis.New(
		rd.NewClient(&rd.Options{
			Addr:     config.GetRedisHost() + ":" + config.GetRedisPort(),
			Password: config.GetRedisPassword(),
			DB:       0,
		}),
	)
	log.Info("Initialized Cache Engine")
}

func (cm *cacheEngine) GetCache() cachego.Cache {
	return cm.cache
}
