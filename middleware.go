package main

import (
	"database/sql"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// JWTMiddleware validates JWT tokens and checks session status
func JWTMiddleware(db *sql.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return SendStandardError(c, ErrorUnauthorized)
			}

			// Check if it starts with "Bearer "
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return SendStandardError(c, ErrorInvalidToken)
			}

			// Extract token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// Check if session is still active
			if !isSessionActive(db, tokenString) {
				return SendStandardError(c, ErrorSessionExpired)
			}

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
				return SendStandardError(c, ErrorInvalidToken)
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

			return SendStandardError(c, ErrorInvalidToken)
		}
	}
}

// isSessionActive checks if the session is still active in database
func isSessionActive(db *sql.DB, token string) bool {
	var isActive bool
	query := `SELECT is_active FROM sessions WHERE token = $1 AND expires_at > NOW()`
	err := db.QueryRow(query, token).Scan(&isActive)
	return err == nil && isActive
}

// getUserIDFromContext extracts user ID from echo context
func getUserIDFromContext(c echo.Context) uuid.UUID {
	if userID, ok := c.Get("user_id").(uuid.UUID); ok {
		return userID
	}
	return uuid.Nil
}
