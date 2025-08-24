package main

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthResult contains the result of JWT authentication
type AuthResult struct {
	UserID primitive.ObjectID
	Email  string
}

// extractUserFromJWT extracts and validates JWT token from Authorization header
// Returns user ID and email if valid, or writes error response and returns nil
func extractUserFromJWT(w http.ResponseWriter, r *http.Request) *AuthResult {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return nil
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return nil
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
		return nil
	}

	email, _ := claims["email"].(string)

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return nil
	}

	return &AuthResult{
		UserID: userID,
		Email:  email,
	}
}
