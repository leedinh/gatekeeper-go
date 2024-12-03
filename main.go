package main

import (
	"context"
	"log"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/leedinh/gatekeeper-go/config"
	"github.com/leedinh/gatekeeper-go/middleware"
	"github.com/redis/go-redis/v9"
)

func main() {
	conf := config.LoadConfig()
	rdb := redis.NewClient(&redis.Options{
		Addr: conf.RedisAddr,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	e := echo.New()

	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())

	// Public routes
	e.GET("/", func(c echo.Context) error {
		return c.String(200, "Hello, World!")
	})

	// Secure routes
	e.GET("/secure", middleware.JWTMiddleware(func(c echo.Context) error {
		return c.String(200, "Secure route")
	}))

	// Limited routes
	e.GET("/limited", middleware.RateLimitMiddleware(func(c echo.Context) error {
		return c.String(200, "Limited route")
	}, rdb))

	// Leaky bucket routes
	e.GET("/leaky", middleware.LeakyBucketMiddleware(func(c echo.Context) error {
		return c.String(200, "Leaky bucket route")
	}))

	log.Printf("Server started at %s\n", conf.ServerPort)
	e.Logger.Fatal(e.Start(":" + conf.ServerPort))
}
