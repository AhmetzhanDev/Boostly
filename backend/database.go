package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GoogleUserInfo представляет информацию о пользователе от Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

var client *mongo.Client
var database *mongo.Database

// ConnectDB подключается к MongoDB
func ConnectDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Подключение к MongoDB (URI из переменной окружения, с дефолтом для dev)
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb+srv://root:root@speakapperai.yaelgbv.mongodb.net/?retryWrites=true&w=majority&appName=SpeakApperAi"
		log.Println("⚠️  MONGODB_URI не задан, используется дефолтный Atlas URI (dev)")
	}
	clientOptions := options.Client().ApplyURI(uri)

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Проверяем подключение
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	database = client.Database("speakapper")
	log.Println("✅ Подключение к MongoDB установлено")
	return nil
}

// DisconnectDB отключается от MongoDB
func DisconnectDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if client != nil {
		return client.Disconnect(ctx)
	}
	return nil
}

// ValidateGoogleToken валидирует Google токен и получает информацию о пользователе
func ValidateGoogleToken(accessToken string) (*GoogleUserInfo, error) {
	url := fmt.Sprintf("https://www.googleapis.com/oauth2/v2/userinfo?access_token=%s", accessToken)
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе к Google API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неверный статус ответа от Google API: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("ошибка при парсинге ответа: %v", err)
	}

	return &userInfo, nil
}

// ValidateGoogleIDToken validates Google ID token and gets user information
func ValidateGoogleIDToken(idToken string) (*GoogleUserInfo, error) {
	url := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error requesting Google tokeninfo API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response status from Google tokeninfo API: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var tokenInfo struct {
		Email         string `json:"email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		EmailVerified string `json:"email_verified"`
		Aud           string `json:"aud"`
	}

	if err := json.Unmarshal(body, &tokenInfo); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	// Verify email is verified
	if tokenInfo.EmailVerified != "true" {
		return nil, fmt.Errorf("email not verified")
	}

	userInfo := &GoogleUserInfo{
		Email:   tokenInfo.Email,
		Name:    tokenInfo.Name,
		Picture: tokenInfo.Picture,
	}

	return userInfo, nil
}

// CreateUser создает нового пользователя в базе данных
func CreateUser(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Генерируем ObjectID для MongoDB
	user.ID = primitive.NewObjectID().Hex()
	user.CreatedAt = time.Now()

	collection := database.Collection("users")
	_, err := collection.InsertOne(ctx, user)
	return err
}

// GetUserByEmail получает пользователя по email
func GetUserByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := database.Collection("users")

	var user User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UserExists проверяет, существует ли пользователь с данным email
func UserExists(email string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := database.Collection("users")

	count, err := collection.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetAllUsers получает всех пользователей (для админки)
func GetAllUsers() ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := database.Collection("users")

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUser обновляет данные пользователя
func UpdateUser(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := database.Collection("users")

	_, err := collection.UpdateOne(
		ctx,
		bson.M{"email": user.Email},
		bson.M{"$set": user},
	)
	return err
}

// DeleteUser удаляет пользователя
func DeleteUser(email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := database.Collection("users")

	_, err := collection.DeleteOne(ctx, bson.M{"email": email})
	return err
}
