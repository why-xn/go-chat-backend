package model

import (
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	ID        uint64         `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type User struct {
	BaseModel
	UserId uint64 `gorm:"index" json:"userId"`
	//Username    string `gorm:"index" json:"username"`
	//FirstName   string `json:"firstName"`
	//LastName    string `json:"lastName"`
	Name           string `json:"username"`
	Type           string `json:"type"`
	Email          string `json:"email,omitempty"`
	PhoneNo        string `json:"phoneNo,omitempty"`
	ProfilePicture string `json:"profilePicture,omitempty"`
	Connections    string `json:"connections,omitempty"`
	LastAuthToken  string `json:"-"`
}

type Connection struct {
	BaseModel
	Uid       string `gorm:"index" json:"uid"`
	UserRefer uint64 `json:"userRefer"`
}

type ChatWindow struct {
	BaseModel
	Uid                        string     `gorm:"index" json:"uid"`
	ParticipantsStr            string     `json:"-"`
	Participants               []User     `gorm:"many2many:chatwindow_participants" json:"participants"`
	LastMessage                string     `gorm:"type:text" json:"lastMessage"`
	LastMessageAt              *time.Time `json:"lastMessageAt" gorm:"default:null"`
	LastMessageSender          string     `json:"lastMessageSender"`
	LastMessageSenderUserId    uint64     `json:"lastMessageSenderUserId"`
	LastMessageSeenByRecipient bool       `json:"lastMessageSeenByRecipient"`
}

type ChatMessage struct {
	BaseModel
	Type            string `json:"type"`
	Message         string `gorm:"type:text" json:"message"`
	ImageUrl        string `json:"imageUrl,omitempty"`
	Sender          uint64 `json:"sender"`
	ChatWindowRefer string `gorm:"index" json:"-"`
}

type ChatSeenStatus struct {
	BaseModel
	ChatWindowRefer uint64 `gorm:"index" json:"chatWindowRefer"`
	UserRefer       uint64 `gorm:"index" json:"userRefer"`
	Seen            bool   `json:"seen"`
}
