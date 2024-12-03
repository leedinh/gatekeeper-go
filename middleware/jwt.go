package middleware

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWTConfig defines the config for JWT middleware.
func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return echo.NewHTTPError(401, "Unauthorized")
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		_, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})

		if err != nil {
			return echo.NewHTTPError(401, fmt.Sprintf("Unauthorized: %v", err))
		}

		return next(c)
	}
}
