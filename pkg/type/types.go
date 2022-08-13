package _type

type Message struct {
	Action       string `json:"action"`
	ChatWindowId string `json:"chatWindowId"`
	ToUser       uint64 `json:"toUser"`
	Message      string `json:"message"`
}

type OutGoingMessage struct {
	Event        string      `json:"event"`
	ChatWindowId string      `json:"chatWindowId,omitempty"`
	From         uint64      `json:"from"`
	ToConnection string      `json:"toConnection,omitempty"`
	Payload      interface{} `json:"payload"`
}

type MessagePayload struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	ImageUrl string `json:"imageUrl,omitempty"`
}

type ChatWindow struct {
	Uid          string
	Participants []string
}

type InitChatWindowInput struct {
	ToUser uint64 `json:"toUser"`
}

type ResponseDTO struct {
	Status  string      `json:"status"`
	Code    string      `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type SendMessageInput struct {
	ConnectionId string `json:"connectionId"`
	Message      string `json:"message"`
}
