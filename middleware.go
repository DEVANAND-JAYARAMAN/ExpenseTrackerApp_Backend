package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// JWTMiddleware validates JWT tokens
func JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{
					Error: "Authorization header required",
				})
			}

			// Check if it starts with "Bearer "
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{
					Error: "Invalid authorization format",
				})
			}

			// Extract token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// Parse and validate token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Validate signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid signing method")
				}

				// Get JWT secret
				secret := os.Getenv("JWT_SECRET")
				if secret == "" {
					secret = "your-secret-key"
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{
					Error: "Invalid or expired token",
				})
			}

			// Extract user ID from claims
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if userIDStr, ok := claims["user_id"].(string); ok {
					if userID, err := uuid.Parse(userIDStr); err == nil {
						// Store user ID in context
						c.Set("user_id", userID)
						return next(c)
					}
				}
			}

			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "Invalid token claims",
			})
		}
	}
}

// getUserIDFromContext extracts user ID from echo context
func getUserIDFromContext(c echo.Context) uuid.UUID {
	if userID, ok := c.Get("user_id").(uuid.UUID); ok {
		return userID
	}
	return uuid.Nil
}