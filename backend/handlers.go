package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// healthHandler handles server health checks
func healthHandler(w http.ResponseWriter, r *http.Request) {
	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"status": "healthy",
		"time":   time.Now(),
	})
}

// signupHandler handles user registration
func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check if user exists
	exists, err := UserExists(req.Email)
	if err != nil {
		log.Printf("Error checking user existence: %v", err)
		JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if exists {
		JSONResponse(w, http.StatusConflict, SignupResponse{
			Success: false,
			Message: "User already exists",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Create user
	user := &User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  string(hashedPassword),
	}

	// Save to database
	if err := CreateUser(user); err != nil {
		log.Printf("Error creating user: %v", err)
		JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Generate JWT token
	token, err := generateJWT(user)
	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Send response
	JSONResponse(w, http.StatusOK, SignupResponse{
		Success: true,
		Message: "User registered successfully",
		User:    user,
		Token:   token,
	})
}

// loginHandler handles user login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user from database
	user, err := GetUserByEmail(req.Email)
	if err != nil {
		JSONError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		JSONError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT token
	token, err := generateJWT(user)
	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Send response
	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Login successful",
		"user":    user,
		"token":   token,
	})
}

// googleSignupHandler handles Google OAuth signup
func googleSignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req GoogleSignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate Google token (handle both access tokens and ID tokens)
	googleUser, err := ValidateGoogleToken(req.Token)
	if err != nil {
		// Try validating as ID token if access token validation fails
		googleUser, err = ValidateGoogleIDToken(req.Token)
		if err != nil {
			log.Printf("Error validating Google token: %v", err)
			JSONError(w, http.StatusUnauthorized, "Invalid Google token")
			return
		}
	}

	// Check if user exists
	exists, err := UserExists(googleUser.Email)
	if err != nil {
		log.Printf("Error checking user existence: %v", err)
		JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var user *User
	if exists {
		// User exists, get their data
		user, err = GetUserByEmail(googleUser.Email)
		if err != nil {
			log.Printf("Error getting existing user: %v", err)
			JSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
	} else {
		// Create new user
		user = &User{
			FirstName: googleUser.GivenName,
			LastName:  googleUser.FamilyName,
			Email:     googleUser.Email,
			Password:  "", // Google users don't have passwords
		}

		if err := CreateUser(user); err != nil {
			log.Printf("Error creating Google user: %v", err)
			JSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
	}

	// Generate JWT token
	token, err := generateJWT(user)
	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Send response
	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Google authentication successful",
		"user":    user,
		"token":   token,
	})
}

// getAllUsersHandler handles getting all users
func getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := GetAllUsers()
	if err != nil {
		JSONErrorWithDetails(w, http.StatusInternalServerError, "Ошибка при получении пользователей", err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"users":   users,
		"count":   len(users),
	})
}

// handleTranscribe handles audio file transcription
func handleTranscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse multipart form (allow large files)
	if err := r.ParseMultipartForm(1024 << 20); err != nil { // 1GB streamed to temp
		JSONError(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		JSONError(w, http.StatusBadRequest, "No audio file provided")
		return
	}
	defer file.Close()

	log.Printf("Received audio file: %s, size: %d bytes", header.Filename, header.Size)

	// Save upload to temp file on disk to avoid memory issues
	tmpDir := os.TempDir()
	tmpIn := filepath.Join(tmpDir, fmt.Sprintf("upload_%d_%s", time.Now().UnixNano(), filepath.Base(header.Filename)))
	out, err := os.Create(tmpIn)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to create temp file")
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		out.Close()
		os.Remove(tmpIn)
		JSONError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	out.Close()
	defer os.Remove(tmpIn)

	// Threshold for long audio (e.g., > 20MB)
	const longThreshold = 20 * 1024 * 1024
	if header.Size > longThreshold {
		text, err := transcribeLongAudio(tmpIn, "")
		if err != nil {
			log.Printf("Long transcription error: %v", err)
			JSONError(w, http.StatusInternalServerError, "Transcription failed")
			return
		}
		JSONResponse(w, http.StatusOK, map[string]interface{}{
			"success":       true,
			"transcription": text,
			"filename":      header.Filename,
			"size":          header.Size,
			"mode":          "segmented",
		})
		return
	}

	// Small file: read and send directly
	fileBytes, err := os.ReadFile(tmpIn)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to read temp file")
		return
	}

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	part, err := writer.CreateFormFile("file", header.Filename)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to create form file")
		return
	}
	part.Write(fileBytes)
	writer.WriteField("model", "whisper-1")
	writer.Close()

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", &requestBody)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to create request")
		return
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("OpenAI API error: %v", err)
		JSONError(w, http.StatusInternalServerError, "Failed to transcribe audio")
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to read response")
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("OpenAI API error: %s - %s", resp.Status, string(respBody))
		JSONError(w, http.StatusInternalServerError, "Transcription failed")
		return
	}

	var whisperResp WhisperResponse
	if err := json.Unmarshal(respBody, &whisperResp); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to parse response")
		return
	}

	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"success":       true,
		"transcription": whisperResp.Text,
		"filename":      header.Filename,
		"size":          header.Size,
		"mode":          "single",
	})
}

// getNoteByID gets a single note by ID with ownership check
func getNoteByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		JSONError(w, http.StatusBadRequest, "ID is required")
		return
	}
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// JWT verification
	auth := extractUserFromJWT(w, r)
	if auth == nil {
		return
	}
	userID := auth.UserID

	coll := client.Database("speakapper").Collection("notes")
	var note Note
	if err := coll.FindOne(context.Background(), bson.M{"_id": objID, "user_id": userID}).Decode(&note); err != nil {
		JSONError(w, http.StatusNotFound, "Not found")
		return
	}
	JSONResponse(w, http.StatusOK, map[string]interface{}{"success": true, "note": note})
}

// getMaterialByID gets a single material by ID with ownership check
func getMaterialByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		JSONError(w, http.StatusBadRequest, "ID is required")
		return
	}
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// JWT verification
	auth := extractUserFromJWT(w, r)
	if auth == nil {
		return
	}
	userID := auth.UserID

	coll := client.Database("speakapper").Collection("materials")
	var mat Material
	if err := coll.FindOne(context.Background(), bson.M{"_id": objID, "user_id": userID}).Decode(&mat); err != nil {
		JSONError(w, http.StatusNotFound, "Not found")
		return
	}

	ff := mat.Flashcards
	if ff == nil {
		ff = []Flashcard{}
	}
	qq := mat.Quiz
	if qq == nil {
		qq = []QuizQuestion{}
	}
	JSONResponse(w, http.StatusOK, map[string]interface{}{"success": true, "material": map[string]interface{}{
		"id":         mat.ID,
		"user_id":    mat.UserID,
		"transcript": mat.Transcript,
		"flashcards": ff,
		"quiz":       qq,
		"created_at": mat.CreatedAt,
		"updated_at": mat.UpdatedAt,
	}})
}

// getUserHandler returns current user data
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	auth := extractUserFromJWT(w, r)
	if auth == nil {
		return
	}
	userID := auth.UserID
	log.Printf("[getUserHandler] Looking for user with ID: %v (type: %T)", userID, userID)

	coll := client.Database("speakapper").Collection("users")
	var user User

	// Try both ObjectID and string formats
	filter := bson.M{"_id": userID}
	err := coll.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		log.Printf("[getUserHandler] User not found with _id filter, trying email: %v", err)
		// Try finding by email if stored in JWT
		if auth.Email != "" {
			filter = bson.M{"email": auth.Email}
			err = coll.FindOne(context.Background(), filter).Decode(&user)
			if err != nil {
				log.Printf("[getUserHandler] User not found by email either: %v", err)
			}
		}
		if err != nil {
			JSONError(w, http.StatusNotFound, "User not found")
			return
		}
	}

	log.Printf("[getUserHandler] Found user: %s %s (%s)", user.FirstName, user.LastName, user.Email)
	// Don't return password hash
	user.Password = ""

	JSONResponse(w, http.StatusOK, user)
}

// deleteNoteByID deletes a note by ID with ownership check
func deleteNoteByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid note ID")
		return
	}

	auth := extractUserFromJWT(w, r)
	if auth == nil {
		return
	}
	userID := auth.UserID

	coll := client.Database("speakapper").Collection("notes")
	result, err := coll.DeleteOne(context.Background(), bson.M{"_id": objID, "user_id": userID})
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to delete note")
		return
	}

	if result.DeletedCount == 0 {
		JSONError(w, http.StatusNotFound, "Note not found")
		return
	}

	JSONResponse(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Note deleted successfully"})
}

// deleteMaterialByID deletes a material by ID with ownership check
func deleteMaterialByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		JSONError(w, http.StatusBadRequest, "ID is required")
		return
	}
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// JWT verification
	auth := extractUserFromJWT(w, r)
	if auth == nil {
		return
	}
	userID := auth.UserID

	coll := client.Database("speakapper").Collection("materials")
	result, err := coll.DeleteOne(context.Background(), bson.M{"_id": objID, "user_id": userID})
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to delete material")
		return
	}
	if result.DeletedCount == 0 {
		JSONError(w, http.StatusNotFound, "Material not found")
		return
	}

	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Material deleted successfully",
	})
}
