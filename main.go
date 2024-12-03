package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/leedinh/gatekeeper-go/config"
	"github.com/leedinh/gatekeeper-go/middleware"
)

func main() {
	conf := config.LoadConfig()
	e := echo.New()

	// Public routes
	e.GET("/", func(c echo.Context) error {
		return c.String(200, "Hello, World!")
	})

	// Secure routes
	e.GET("/secure", middleware.JWTMiddleware(func(c echo.Context) error {
		return c.String(200, "Secure route")
	}))

	// Limited routes
	e.GET("/limited", func(c echo.Context) error {
		return c.String(200, "Limited route")
	})

	log.Printf("Server started at %s\n", conf.ServerPort)
	e.Logger.Fatal(e.Start(":" + conf.ServerPort))
}
