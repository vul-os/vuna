package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"firebase.google.com/go/auth"
)

// User represents the authenticated user
type User struct {
	ID           string
	Email        string
	DisplayName  string
	PhotoURL     string
	// Add more fields as needed
}

// Middleware for authentication
func IsAuthenticated(authClient *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		idToken := c.GetHeader("Authorization")
		if idToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		token, err := authClient.VerifyIDToken(c, idToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Create User struct with user information
		user := User{
			ID:          token.UID,
			Email:       token.Claims["email"].(string),
			DisplayName: token.Claims["name"].(string),
			PhotoURL:    token.Claims["picture"].(string),
			// Add more fields as needed
		}

		// Store the authenticated user in the context
		c.Set("user", user)

		c.Next()
	}
}
