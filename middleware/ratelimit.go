package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/leedinh/gatekeeper-go/config"
	"github.com/redis/go-redis/v9"
)

// TokenBucket defines the token bucket for rate limiting.
func RateLimitMiddleware(next echo.HandlerFunc, rdb *redis.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		conf := config.LoadConfig()

		// Get ip of user
		ip := c.RealIP()

		key := fmt.Sprintf("rate_limit:%s", ip)

		// Get current count
		count, err := rdb.Get(context.Background(), key).Int()
		if err != nil && err != redis.Nil {
			return echo.NewHTTPError(500, "Internal server error")
		}

		if count >= conf.RateLimit {
			return echo.NewHTTPError(429, "Too many requests")
		}

		// Increment count
		err = rdb.Incr(context.Background(), key).Err()
		if err != nil {
			return echo.NewHTTPError(500, "Internal server error")
		}

		err = rdb.Expire(context.Background(), key, time.Duration(conf.RateLimitTTl)*time.Second).Err()
		if err != nil {
			return echo.NewHTTPError(500, "Internal server error")
		}

		return next(c)

	}
}
