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

// LeakyBucketMiddleware applies the leaky bucket algorithm for rate limiting.
func LeakyBucketMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the config from the context
		conf := c.Get("config").(config.Config)

		// Get the IP address of the user making the request
		ip := c.RealIP()

		// Create a Redis client
		rdb := redis.NewClient(&redis.Options{
			Addr: conf.RedisAddr, // Redis server address from the config
		})

		// Define the key for the user's leaky bucket
		key := fmt.Sprintf("leaky_bucket:%s", ip)

		// Get the current state of the bucket (number of requests and last timestamp)
		bucketState, err := rdb.HGetAll(context.Background(), key).Result()
		if err != nil && err != redis.Nil {
			// If there is an error accessing Redis (e.g., Redis is down), return internal server error
			return echo.NewHTTPError(500, "Internal Server Error")
		}

		// If the bucket doesn't exist yet, initialize it
		if len(bucketState) == 0 {
			// Set the initial state: 0 requests, last processed at current time
			rdb.HSet(context.Background(), key, "count", 0, "last_processed", time.Now().Unix())
			bucketState = map[string]string{
				"count":          "0",
				"last_processed": fmt.Sprintf("%d", time.Now().Unix()),
			}
		}

		// Extract the current count and the timestamp of the last processed request
		count := bucketState["count"]
		lastProcessed := bucketState["last_processed"]

		// Parse values
		currentCount := 0
		if count != "" {
			currentCount = int(count[0] - '0') // Convert count to integer
		}

		lastTime, _ := time.Parse("Unix", lastProcessed)
		timeElapsed := time.Now().Sub(lastTime).Seconds()

		// Calculate the number of requests that should have leaked out by now
		leakedCount := int(timeElapsed / float64(conf.LeakRate)) // LeakRate is in seconds per request

		// Update the count in the bucket
		if currentCount > leakedCount {
			// Decrease the count by the leaked requests
			currentCount -= leakedCount
		} else {
			// If there wasn't enough time for any requests to leak, reset the count
			currentCount = 0
		}

		// Check if the user has exceeded the bucket capacity
		if currentCount >= conf.LeakyBucketCapacity {
			return echo.NewHTTPError(429, "Too Many Requests")
		}

		// Increment the bucket count and update the last processed time
		err = rdb.HSet(context.Background(), key, "count", currentCount+1, "last_processed", time.Now().Unix()).Err()
		if err != nil {
			return echo.NewHTTPError(500, "Internal Server Error")
		}

		// Proceed to the next handler
		return next(c)
	}
}
