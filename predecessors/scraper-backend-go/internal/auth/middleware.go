package auth

import (
	"context"
	"net/http"

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
func IsAuthenticated(authClient *auth.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idToken := r.Header.Get("Authorization")
			if idToken == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			token, err := authClient.VerifyIDToken(context.Background(), idToken)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
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

			// Store the authenticated user in the request context
			ctx := context.WithValue(r.Context(), "user", user)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
