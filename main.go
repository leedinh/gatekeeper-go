package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/leedinh/gatekeeper-go/middleware"
)

func main() {
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

	log.Println("Server started at :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
