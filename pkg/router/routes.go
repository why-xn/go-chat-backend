package router

import (
	"github.com/labstack/echo"
	apiV1 "github.com/whyxn/go-chat-backend/pkg/api/v1"
	"net/http"
)

func Routes(e *echo.Echo) {

	apiV1Chat := apiV1.NewChatApi()

	// Index Page
	e.GET("/", index)

	e.GET("/api/v1/users", apiV1Chat.GetAllUsers)

	e.GET("/api/v1/users/myself", apiV1Chat.GetMe)

	e.GET("/api/v1/chat/:id", apiV1Chat.GetChatWindow)

	e.POST("/api/v1/chat/init", apiV1Chat.InitChatWindow)

	e.GET("/api/v1/chat/windows", apiV1Chat.GetChatWindowList)

	e.POST("/api/v1/chat/:id/messages", apiV1Chat.SendMessage)

	e.GET("/api/v1/chat/:id/messages", apiV1Chat.GetMessages)

}

func index(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{})
}
