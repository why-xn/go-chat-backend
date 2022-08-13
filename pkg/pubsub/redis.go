package pubsub

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/whyxn/go-chat-backend/pkg/config"
	"gopkg.in/redis.v4"
	"net"
	"reflect"
)

type RedisPubSub struct {
	client *redis.Client
}

var Service *RedisPubSub

func init() {
	config.LoadEnvironmentVariables()

	log.Info("Redis pubsub client init...")
	var client *redis.Client
	client = redis.NewClient(&redis.Options{
		Addr:     config.GetRedisHost() + ":" + config.GetRedisPort(),
		Password: config.GetRedisPassword(),
		DB:       1,
		PoolSize: 10,
	})
	Service = &RedisPubSub{client}

	_, err := client.Ping().Result()

	if err != nil {
		log.Fatal("Error connecting to Redis PubSub", err)
	}
}

func (rps *RedisPubSub) PublishString(channel, message string) *redis.IntCmd {
	return rps.client.Publish(channel, message)
}

func (rps *RedisPubSub) Publish(channel string, message interface{}) *redis.IntCmd {
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}
	messageString := string(jsonBytes)
	return rps.client.Publish(channel, messageString)
}

type Subscriber struct {
	pubsub   *redis.PubSub
	channel  string
	callback processFunc
}

type processFunc func(string, string)

func NewSubscriber(channel string, fn processFunc) (*Subscriber, error) {
	var err error

	s := Subscriber{
		channel:  channel,
		callback: fn,
	}

	// Subscribe to the channel
	err = s.subscribe()
	if err != nil {
		return nil, err
	} else {
		return nil, err
	}

	// Listen for messages
	go s.listen()

	return &s, nil
}

func (s *Subscriber) subscribe() error {
	var err error
	s.pubsub, err = Service.client.Subscribe(s.channel)
	if err != nil {
		log.Error("Error subscribing to channel.")
		return err
	}
	return nil
}

func (s *Subscriber) listen() error {
	for {
		msg, err := s.pubsub.ReceiveMessage()
		if err != nil {
			if reflect.TypeOf(err) == reflect.TypeOf(&net.OpError{}) && (reflect.TypeOf(err.(*net.OpError).Err).String() == "*net.timeoutError" || reflect.TypeOf(err.(*net.OpError).Err).String() == "*poll.TimeoutError") {
				// Timeout, ignore
				continue
			} // Actual error
			log.Error("Error in ReceiveMessage()", err)
		}

		// Process the message
		if len(msg.Channel) > 0 {
			go s.callback(msg.Channel, msg.Payload)
		}
	}
}
