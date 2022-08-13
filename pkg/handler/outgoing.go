package handler

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/whyxn/go-chat-backend/pkg/core"
	"github.com/whyxn/go-chat-backend/pkg/type"
)

func OutGoingHandler(channel, payload string) {
	if len(payload) > 0 {
		var om _type.OutGoingMessage

		err := json.Unmarshal([]byte(payload), &om)
		if err != nil {
			log.Error("Unmarshal error: %v", err)
			return
		}

		if len(om.ToConnection) > 0 {
			if _, ok := core.GetChatEngine().GetGoChat().ClientsMap[om.ToConnection]; ok {
				client := core.GetChatEngine().GetGoChat().ClientsMap[om.ToConnection]
				client.Send([]byte(payload))
			}
		}
	}
}
