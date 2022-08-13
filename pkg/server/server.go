package server

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/whyxn/go-chat-backend/pkg/core"
	"html/template"
)

func New() *echo.Echo {
	echoInstance := echo.New()
	echoInstance.Use(middleware.Recover())

	renderer := &core.TemplateRenderer{
		Templates: template.Must(template.ParseGlob("public/views/*.html")),
	}
	echoInstance.Renderer = renderer

	// Configuring Middleware Logger
	echoInstance.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		// Skipping logging for health checking api
		Skipper: func(c echo.Context) bool {
			if c.Request().RequestURI == "/health" {
				return true
			}
			return false
		},
		Format: "[${time_rfc3339}] method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	}))

	echoInstance.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	return echoInstance
}
