package core

import (
	log "github.com/sirupsen/logrus"
	"time"
)

var err error

func PingAllClients() {

	for _, client := range GetChatEngine().GetGoChat().ClientsMap {
		err = client.Send([]byte("p"))
		if err != nil {
			log.Error("Error occurred while pinging to client", client.Id, ". Msg:", err.Error())

			time.Sleep(2 * time.Second)

			err = client.Send([]byte("p"))
			if err != nil {
				GetChatEngine().GetGoChat().RemoveClient(client)
			}

		}
	}

	time.Sleep(10 * time.Second)

	PingAllClients()

}
