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
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// User представляет структуру пользователя
type User struct {
	ID        string    `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Не отправляем пароль в JSON
	CreatedAt time.Time `json:"createdAt"`
}

// Transcribe YouTube by URL using yt-dlp + Whisper
func handleTranscribeYouTube(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Expect JSON: {"url": "https://youtu.be/..."}
	var body struct {
		URL      string `json:"url"`
		Language string `json:"language,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("YouTube transcribe: invalid body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if body.URL == "" {
		log.Printf("YouTube transcribe: missing url")
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}
	log.Printf("YouTube transcribe: start url=%s", body.URL)

	// Check yt-dlp availability
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		log.Printf("yt-dlp not found: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusFailedDependency)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "yt-dlp is required on server",
			"hint":    "Install with: brew install yt-dlp (mac) or pipx install yt-dlp",
		})
		return
	}

	// Prepare temp paths
	tmpDir := os.TempDir()
	base := fmt.Sprintf("yt_%d", time.Now().UnixNano())
	outPath := filepath.Join(tmpDir, base+".mp3")

	// Download audio with robust format preferences and fallbacks to avoid m3u8 403
	// 1) Prefer non-m3u8 m4a to avoid HLS fragment 403
	// 2) Fallback to generic bestaudio/best
	// 3) Try using cookies from Chrome/Safari (mac) to bypass geo/age restrictions
	ytdlpOutput := func(args ...string) ([]byte, error) {
		cmd := exec.Command("yt-dlp", args...)
		return cmd.CombinedOutput()
	}

	outPattern := filepath.Join(tmpDir, base+".%(ext)s")
	// Optional cookies.txt from env to avoid browser permissions
	cookiesArgs := []string{}
	if cp := os.Getenv("YTDLP_COOKIES"); cp != "" {
		if _, err := os.Stat(cp); err == nil {
			cookiesArgs = []string{"--cookies", cp}
			log.Printf("yt-dlp will use cookies file: %s", cp)
		} else {
			log.Printf("cookies file not found at YTDLP_COOKIES=%s (ignored)", cp)
		}
	}

	// Define a helper to try different player clients without cookies
	tryClient := func(clientName string, format string) ([]byte, error) {
		args := append([]string{"-R", "3", "--fragment-retries", "3", "--force-ipv4", "--geo-bypass", "--extractor-args", fmt.Sprintf("youtube:player_client=%s", clientName), "-f", format, "-x", "--audio-format", "mp3", "-o", outPattern}, cookiesArgs...)
		args = append(args, body.URL)
		log.Printf("yt-dlp try-client=%s args=%v", clientName, args)
		return ytdlpOutput(args...)
	}

	// attempt 1: web client with strict non-HLS preference
	outBytes, err := tryClient("web", "bestaudio[ext=m4a]/bestaudio[protocol!=m3u8]/bestaudio/best")
	if err != nil {
		log.Printf("yt-dlp attempt1 (web) failed: %v; output: %s", err, string(outBytes))
		// attempt 2: web simpler
		outBytes2, err2 := tryClient("web", "bestaudio/best")
		if err2 != nil {
			log.Printf("yt-dlp attempt2 (web simple) failed: %v; output: %s", err2, string(outBytes2))
			// attempt 3: android client
			outBytesA, errA := tryClient("android", "bestaudio[ext=m4a]/bestaudio/best")
			if errA != nil {
				log.Printf("yt-dlp attempt3 (android) failed: %v; output: %s", errA, string(outBytesA))
				// attempt 4: ios client
				outBytesI, errI := tryClient("ios", "bestaudio[ext=m4a]/bestaudio/best")
				if errI != nil {
					log.Printf("yt-dlp attempt4 (ios) failed: %v; output: %s", errI, string(outBytesI))
					// attempt 5: tvhtml5 client
					outBytesT, errT := tryClient("tvhtml5", "bestaudio[ext=m4a]/bestaudio/best")
					if errT != nil {
						log.Printf("yt-dlp attempt5 (tvhtml5) failed: %v; output: %s", errT, string(outBytesT))
						// attempt 6: optional chrome cookies if available (no Safari to avoid permissions)
						args3c := []string{"--cookies-from-browser", "chrome", "-R", "3", "--fragment-retries", "3", "--force-ipv4", "--geo-bypass", "--extractor-args", "youtube:player_client=web", "-f", "bestaudio[ext=m4a]/bestaudio/best", "-x", "--audio-format", "mp3", "-o", outPattern, body.URL}
						log.Printf("yt-dlp retry (with Chrome cookies): args=%v", args3c)
						outBytes3c, err3c := ytdlpOutput(args3c...)
						if err3c != nil {
							log.Printf("yt-dlp chrome-cookies failed: %v; output: %s", err3c, string(outBytes3c))
							w.Header().Set("Content-Type", "application/json")
							w.WriteHeader(http.StatusConflict)
							json.NewEncoder(w).Encode(map[string]interface{}{
								"success": false,
								"message": "Не удалось скачать аудио с YouTube без авторизации",
								"details": string(outBytes3c),
								"hint":    "Для гарантии работы: загрузите файл (видео/аудио) напрямую или укажите YTDLP_COOKIES=путь/к/cookies.txt.",
							})
							return
						}
					}
				}
			}
		}
	}

	// Determine produced file. We expect mp3 with base.mp3
	if _, err := os.Stat(outPath); err != nil {
		// try to find any produced file
		entries, _ := os.ReadDir(tmpDir)
		found := ""
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), base+".") {
				found = filepath.Join(tmpDir, e.Name())
				break
			}
		}
		if found == "" {
			log.Printf("YouTube transcribe: audio not found after yt-dlp, base=%s", base)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Audio file not found after download",
			})
			return
		}
		outPath = found
	}
	defer os.Remove(outPath)

	// Всегда используем сегментированную транскрипцию для YouTube
	info, _ := os.Stat(outPath)
	if info != nil {
		log.Printf("YouTube transcribe: downloaded file=%s size=%d bytes", outPath, info.Size())
	}
	text, err := transcribeLongAudio(outPath, body.Language)
	if err != nil {
		log.Printf("YouTube segmented transcription error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Transcription failed",
			"error":   err.Error(),
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"transcription": text,
		"source":        "youtube",
		"url":           body.URL,
		"mode":          "segmented",
	})

	log.Printf("YouTube transcribe: success segmented url=%s", body.URL)
}

// Генерация и сохранение материалов в одну операцию
func handleGenerateAndSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Method not allowed"})
		return
	}

	// Debug: вход в обработчик
	log.Println("[handleGenerateAndSave] start")

	// Извлекаем пользователя из JWT
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Authorization header required"})
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		log.Printf("[handleGenerateAndSave] invalid token: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Invalid token"})
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("[handleGenerateAndSave] invalid token claims type")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Invalid token claims"})
		return
	}
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		// Иногда ObjectID может прийти не строкой — логируем тип для диагностики
		log.Printf("[handleGenerateAndSave] user_id claim not string, actual=%T value=%v", claims["user_id"], claims["user_id"])
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Invalid user ID in token"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		log.Printf("[handleGenerateAndSave] invalid userID hex: %s error=%v", userIDStr, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Invalid user ID format"})
		return
	}
	log.Printf("[handleGenerateAndSave] userID=%s", userID.Hex())

	// Читаем тело запроса
	var reqBody GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("[handleGenerateAndSave] decode body error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Invalid request body"})
		return
	}
	if reqBody.Transcript == "" {
		log.Println("[handleGenerateAndSave] empty transcript")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Transcript is required"})
		return
	}

	// Подсчёт желаемого числа вопросов от объёма текста
	words := len(strings.Fields(reqBody.Transcript))
	targetQuiz := 0
	if words > 0 {
		targetQuiz = words / 120 // ~1 вопрос на 120 слов
		if targetQuiz < 6 {
			targetQuiz = 6
		}
		if targetQuiz > 30 {
			targetQuiz = 30
		}
	}
	log.Printf("[handleGenerateAndSave] transcript_len=%d words=%d targetQuiz~=%d", len(reqBody.Transcript), words, targetQuiz)

	// Усиленный промпт с логикой квизов и балансом типов
	systemPrompt := strings.Join([]string{
		"You convert transcripts into study materials.",
		"RULES:",
		// ...
	}, "\n")

	userPrompt := fmt.Sprintf(
		"Language hint: %s\nAim for ~%d total quiz questions given transcript length (adjust down if insufficient material).\nTranscript:\n%s\n\nReturn JSON only.",
		reqBody.Language, targetQuiz, reqBody.Transcript,
	)

	chatReq := map[string]interface{}{
		"model":           "gpt-4o-mini",
		"temperature":     0.3,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}
	buf, _ := json.Marshal(chatReq)
	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(buf))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 70 * time.Second}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		log.Printf("[handleGenerateAndSave] OpenAI chat API error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Failed to generate materials", "error": err.Error()})
		return
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Failed to read response"})
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("[handleGenerateAndSave] OpenAI chat API error: %s - %s", resp.Status, string(respBytes))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Generation failed", "status": resp.Status})
		return
	}

	var openaiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBytes, &openaiResp); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Failed to parse OpenAI response"})
		return
	}
	if len(openaiResp.Choices) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Empty OpenAI response"})
		return
	}

	var payload GeneratePayload
	if err := json.Unmarshal([]byte(openaiResp.Choices[0].Message.Content), &payload); err != nil {
		clean := strings.TrimSpace(openaiResp.Choices[0].Message.Content)
		clean = strings.TrimPrefix(clean, "```json")
		clean = strings.TrimPrefix(clean, "```")
		clean = strings.TrimSuffix(clean, "```")
		clean = strings.TrimSpace(clean)
		if err2 := json.Unmarshal([]byte(clean), &payload); err2 != nil {
			log.Printf("[handleGenerateAndSave] Failed to unmarshal AI JSON: %v; content: %s", err, openaiResp.Choices[0].Message.Content)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Invalid JSON from model"})
			return
		}
	}

	// Сохраняем материал в MongoDB с привязкой к пользователю
	material := Material{
		UserID:     userID,
		Transcript: reqBody.Transcript,
		Flashcards: payload.Flashcards,
		Quiz:       payload.Quiz,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	collection := client.Database("speakapper").Collection("materials")
	result, err := collection.InsertOne(context.Background(), material)
	if err != nil {
		log.Printf("[handleGenerateAndSave] Error saving material: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Failed to save material"})
		return
	}
	material.ID = result.InsertedID.(primitive.ObjectID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"material": material,
		// дублируем для удобства фронтенда
		"flashcards": material.Flashcards,
		"quiz":       material.Quiz,
	})
}

// Получение одной заметки по ID (с проверкой владельца)
func getNoteByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Проверка JWT
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) { return jwtSecret, nil })
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	userIDStr, _ := claims["user_id"].(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	coll := client.Database("speakapper").Collection("notes")
	var note Note
	if err := coll.FindOne(context.Background(), bson.M{"_id": objID, "user_id": userID}).Decode(&note); err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "note": note})
}

// Получение одного материала по ID (с проверкой владельца)
func getMaterialByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Проверка JWT
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) { return jwtSecret, nil })
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	userIDStr, _ := claims["user_id"].(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	coll := client.Database("speakapper").Collection("materials")
	var mat Material
	if err := coll.FindOne(context.Background(), bson.M{"_id": objID, "user_id": userID}).Decode(&mat); err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "material": mat})
}

// SignupRequest представляет запрос на регистрацию
type SignupRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// SignupResponse представляет ответ на регистрацию
type SignupResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	User    *User  `json:"user,omitempty"`
	Token   string `json:"token,omitempty"`
}

// LoginRequest представляет запрос на вход
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// GoogleSignupRequest представляет запрос на регистрацию через Google
type GoogleSignupRequest struct {
	Token   string `json:"token"`
	IDToken string `json:"idToken"`
}

// JWT секретный ключ (в продакшене используйте переменную окружения)
var jwtSecret = []byte("your-secret-key")

// OpenAI API ключ теперь берём из переменной окружения
var openaiAPIKey string

// OpenAI API структуры
type WhisperRequest struct {
	Model    string `json:"model"`
	File     string `json:"file"`
	Language string `json:"language,omitempty"`
}

type WhisperResponse struct {
	Text string `json:"text"`
}

// Note структура для MongoDB
type Note struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	Title      string             `bson:"title" json:"title"`
	Content    string             `bson:"content" json:"content"`
	Type       string             `bson:"type" json:"type"`
	Tab        string             `bson:"tab" json:"tab"`
	LastOpened string             `bson:"last_opened" json:"last_opened"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

// GPT generation types
type GenerateRequest struct {
	Transcript string `json:"transcript"`
	Language   string `json:"language,omitempty"`
}

type Flashcard struct {
	Term       string `json:"term"`
	Definition string `json:"definition"`
	Example    string `json:"example,omitempty"`
}

// FlexString позволяет распаковывать как строки, так и числа в строковое поле
type FlexString string

func (s *FlexString) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	// Строка в кавычках
	if b[0] == '"' {
		var str string
		if err := json.Unmarshal(b, &str); err != nil {
			return err
		}
		*s = FlexString(str)
		return nil
	}
	// Число -> строка
	var num json.Number
	if err := json.Unmarshal(b, &num); err == nil {
		*s = FlexString(num.String())
		return nil
	}
	// Фоллбэк: любое значение -> строка
	var v interface{}
	if err := json.Unmarshal(b, &v); err == nil {
		*s = FlexString(fmt.Sprint(v))
		return nil
	}
	return fmt.Errorf("invalid value for FlexString: %s", string(b))
}

type QuizQuestion struct {
	ID         FlexString `json:"id,omitempty"`
	Type       string     `json:"type,omitempty"` // MCQ, MSQ, CLOZE, TF, MATCHING, SHORT
	Question   string     `json:"question"`
	Options    []string   `json:"options,omitempty"`    // MCQ/MSQ/TF/CLOZE
	Answer     string     `json:"answer,omitempty"`     // Верный ответ для MCQ/TF/SHORT/CLOZE
	Correct    []string   `json:"correct,omitempty"`    // Верные ответы для MSQ
	Pairs      [][]string `json:"pairs,omitempty"`      // MATCHING: массив пар [[left,right], ...]
	Rationale  string     `json:"rationale,omitempty"`  // Объяснение
	Difficulty string     `json:"difficulty,omitempty"` // easy|medium|hard
	Citation   string     `json:"citation,omitempty"`   // Цитата/ссылка на фрагмент транскрипта
}

type GeneratePayload struct {
	Flashcards   []Flashcard    `json:"flashcards"`
	Quiz         []QuizQuestion `json:"quiz"`
	LanguageCode string         `json:"languageCode,omitempty"`
}

// Учебные материалы (материализованные карточки/квиз)
type Material struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	Transcript string             `bson:"transcript" json:"transcript"`
	Flashcards []Flashcard        `bson:"flashcards" json:"flashcards"`
	Quiz       []QuizQuestion     `bson:"quiz" json:"quiz"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

func handleMaterials(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// Handle GET request - fetch user materials
	if r.Method == "GET" {
		collection := client.Database("speakapper").Collection("materials")

		// Find all materials for this user
		cursor, err := collection.Find(context.Background(), bson.M{"user_id": userID})
		if err != nil {
			log.Printf("Error fetching materials: %v", err)
			http.Error(w, "Failed to fetch materials", http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		var materials []Material
		if err = cursor.All(context.Background(), &materials); err != nil {
			log.Printf("Error decoding materials: %v", err)
			http.Error(w, "Failed to decode materials", http.StatusInternalServerError)
			return
		}

		// Convert ObjectIDs to strings for JSON response
		var responseMaterials []map[string]interface{}
		for _, mat := range materials {
			responseMaterials = append(responseMaterials, map[string]interface{}{
				"id":         mat.ID.Hex(),
				"transcript": mat.Transcript,
				"flashcards": mat.Flashcards,
				"quiz":       mat.Quiz,
				"created_at": mat.CreatedAt,
				"updated_at": mat.UpdatedAt,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"materials": responseMaterials,
		})
		return
	}

	// Handle POST request - create new material
	var materialData struct {
		Transcript string         `json:"transcript"`
		Flashcards []Flashcard    `json:"flashcards"`
		Quiz       []QuizQuestion `json:"quiz"`
	}

	if err := json.NewDecoder(r.Body).Decode(&materialData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create material
	material := Material{
		UserID:     userID,
		Transcript: materialData.Transcript,
		Flashcards: materialData.Flashcards,
		Quiz:       materialData.Quiz,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save to MongoDB
	collection := client.Database("speakapper").Collection("materials")
	result, err := collection.InsertOne(context.Background(), material)
	if err != nil {
		log.Printf("Error saving material: %v", err)
		http.Error(w, "Failed to save material", http.StatusInternalServerError)
		return
	}

	// Set the ID from the inserted document
	material.ID = result.InsertedID.(primitive.ObjectID)

	// Return the created material
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"material": material,
	})
}

func handleNotes(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// Handle GET request - fetch user notes
	if r.Method == "GET" {
		collection := client.Database("speakapper").Collection("notes")

		// Find all notes for this user
		cursor, err := collection.Find(context.Background(), bson.M{"user_id": userID})
		if err != nil {
			log.Printf("Error fetching notes: %v", err)
			http.Error(w, "Failed to fetch notes", http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		var notes []Note
		if err = cursor.All(context.Background(), &notes); err != nil {
			log.Printf("Error decoding notes: %v", err)
			http.Error(w, "Failed to decode notes", http.StatusInternalServerError)
			return
		}

		// Convert ObjectIDs to strings for JSON response
		var responseNotes []map[string]interface{}
		for _, note := range notes {
			responseNotes = append(responseNotes, map[string]interface{}{
				"id":          note.ID.Hex(),
				"title":       note.Title,
				"content":     note.Content,
				"type":        note.Type,
				"tab":         note.Tab,
				"last_opened": note.LastOpened,
				"created_at":  note.CreatedAt,
				"updated_at":  note.UpdatedAt,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"notes":   responseNotes,
		})
		return
	}

	// Handle POST request - create new note
	// Parse request body
	var noteData struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Type    string `json:"type"`
		Tab     string `json:"tab"`
	}

	if err := json.NewDecoder(r.Body).Decode(&noteData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create note
	note := Note{
		UserID:     userID,
		Title:      noteData.Title,
		Content:    noteData.Content,
		Type:       noteData.Type,
		Tab:        noteData.Tab,
		LastOpened: "Just now",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save to MongoDB
	collection := client.Database("speakapper").Collection("notes")
	result, err := collection.InsertOne(context.Background(), note)
	if err != nil {
		log.Printf("Error saving note: %v", err)
		http.Error(w, "Failed to save note", http.StatusInternalServerError)
		return
	}

	// Set the ID from the inserted document
	note.ID = result.InsertedID.(primitive.ObjectID)

	// Return the created note
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"note":    note,
	})
}

func handleTranscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (allow large)
	if err := r.ParseMultipartForm(1024 << 20); err != nil { // 1GB streamed to temp
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "No audio file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Received audio file: %s, size: %d bytes", header.Filename, header.Size)

	// Save upload to temp file on disk to avoid memory blowups
	tmpDir := os.TempDir()
	tmpIn := filepath.Join(tmpDir, fmt.Sprintf("upload_%d_%s", time.Now().UnixNano(), filepath.Base(header.Filename)))
	out, err := os.Create(tmpIn)
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		out.Close()
		os.Remove(tmpIn)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
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
			http.Error(w, "Transcription failed", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
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
		http.Error(w, "Failed to read temp file", http.StatusInternalServerError)
		return
	}

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	part, err := writer.CreateFormFile("file", header.Filename)
	if err != nil {
		http.Error(w, "Failed to create form file", http.StatusInternalServerError)
		return
	}
	part.Write(fileBytes)
	writer.WriteField("model", "whisper-1")
	writer.Close()

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", &requestBody)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("OpenAI API error: %v", err)
		http.Error(w, "Failed to transcribe audio", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("OpenAI API error: %s - %s", resp.Status, string(respBody))
		http.Error(w, "Transcription failed", http.StatusInternalServerError)
		return
	}

	var whisperResp WhisperResponse
	if err := json.Unmarshal(respBody, &whisperResp); err != nil {
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"transcription": whisperResp.Text,
		"filename":      header.Filename,
		"size":          header.Size,
		"mode":          "single",
	})
}

func transcribeLongAudio(inputPath string, language string) (string, error) {
	// Check ffmpeg availability
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return "", fmt.Errorf("ffmpeg not found: %w", err)
	}

	workDir, err := os.MkdirTemp(os.TempDir(), "chunks_*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(workDir)

	// Segment into ~10-minute chunks, mono 16kHz low bitrate to reduce size
	chunkPattern := filepath.Join(workDir, "chunk_%03d.mp3")
	cmd := exec.Command("ffmpeg", "-y", "-i", inputPath, "-ac", "1", "-ar", "16000", "-b:a", "64k", "-f", "segment", "-segment_time", "600", chunkPattern)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg segment failed: %w", err)
	}

	// Collect chunks
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return "", err
	}
	var chunkFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "chunk_") {
			chunkFiles = append(chunkFiles, filepath.Join(workDir, e.Name()))
		}
	}
	if len(chunkFiles) == 0 {
		return "", fmt.Errorf("no chunks produced")
	}
	// Sort by name to keep order
	slices.Sort(chunkFiles)

	var fullText strings.Builder
	httpClient := &http.Client{Timeout: 180 * time.Second}
	for _, path := range chunkFiles {
		// Build multipart per chunk
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		chunk, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		p, err := mw.CreateFormFile("file", filepath.Base(path))
		if err != nil {
			return "", err
		}
		if _, err := p.Write(chunk); err != nil {
			return "", err
		}
		mw.WriteField("model", "whisper-1")
		if language != "" {
			mw.WriteField("language", language)
		}
		mw.Close()

		req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", &body)
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
		req.Header.Set("Content-Type", mw.FormDataContentType())

		resp, err := httpClient.Do(req)
		if err != nil {
			return "", err
		}
		respBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("chunk transcribe error: %s - %s", resp.Status, string(respBytes))
		}
		var wr WhisperResponse
		if err := json.Unmarshal(respBytes, &wr); err != nil {
			return "", err
		}
		fullText.WriteString(strings.TrimSpace(wr.Text))
		fullText.WriteString("\n")
		// throttle a bit to be safe
		time.Sleep(500 * time.Millisecond)
	}

	return fullText.String(), nil
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if reqBody.Transcript == "" {
		http.Error(w, "Transcript is required", http.StatusBadRequest)
		return
	}

	// Build chat completion request
	chatReq := map[string]interface{}{
		"model":           "gpt-4o-mini",
		"temperature":     0.3,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": "You convert transcripts into study materials. IMPORTANT RULES: 1) USE ONLY words and facts from the transcript; DO NOT invent. 2) Flashcards: term must be an exact word/phrase from transcript; definition should be a short sentence fragment from transcript (or closest sentence). 3) Quiz: each question must be based on transcript; the correct answer and ALL options must be text spans that appear in the transcript. 4) If there is not enough material, return fewer items. Respond strictly as compact JSON with keys: flashcards (array of {term, definition, example?}) and quiz (array of {id?, topicId?, question, options[4], answer, hint?}). No extra commentary, no markdown fences."},
			{"role": "user", "content": fmt.Sprintf("Language hint: %s\nTranscript:\n%s\n\nReturn JSON only.", reqBody.Language, reqBody.Transcript)},
		},
	}

	buf, _ := json.Marshal(chatReq)
	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(buf))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 45 * time.Second}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		log.Printf("OpenAI chat API error: %v", err)
		http.Error(w, "Failed to generate materials", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("OpenAI chat API error: %s - %s", resp.Status, string(respBytes))
		http.Error(w, "Generation failed", http.StatusInternalServerError)
		return
	}

	// Parse choices[0].message.content
	var openaiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBytes, &openaiResp); err != nil {
		http.Error(w, "Failed to parse OpenAI response", http.StatusInternalServerError)
		return
	}
	if len(openaiResp.Choices) == 0 {
		http.Error(w, "Empty OpenAI response", http.StatusInternalServerError)
		return
	}

	var payload GeneratePayload
	if err := json.Unmarshal([]byte(openaiResp.Choices[0].Message.Content), &payload); err != nil {
		// If model returned fenced code block, try strip
		clean := openaiResp.Choices[0].Message.Content
		clean = strings.TrimSpace(clean)
		// naive cleanup of Markdown fences
		clean = strings.TrimPrefix(clean, "```json")
		clean = strings.TrimPrefix(clean, "```")
		clean = strings.TrimSuffix(clean, "```")
		if err2 := json.Unmarshal([]byte(clean), &payload); err2 != nil {
			log.Printf("Failed to unmarshal AI JSON: %v; content: %s", err, openaiResp.Choices[0].Message.Content)
			http.Error(w, "Invalid JSON from model", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"flashcards": payload.Flashcards,
		"quiz":       payload.Quiz,
	})
}

func main() {
	// Читаем OpenAI API ключ из переменной окружения
	openaiAPIKey = os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		log.Fatal("❌ OPENAI_API_KEY не задан в переменных окружения!")
	}

	// Читаем JWT секрет из окружения (с дефолтом и предупреждением)
	if envSecret := os.Getenv("JWT_SECRET"); envSecret != "" {
		jwtSecret = []byte(envSecret)
	} else {
		log.Println("⚠️  JWT_SECRET не задан, используется небезопасный дефолтный ключ. Задайте JWT_SECRET в окружении!")
	}

	// Подключаемся к MongoDB
	if err := ConnectDB(); err != nil {
		log.Fatal("❌ Ошибка подключения к MongoDB:", err)
	}
	defer DisconnectDB()
	r := mux.NewRouter()

	// Настройка CORS
	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{
			"http://localhost:3001",
			"http://localhost:3000",
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)

	// Роуты
	r.HandleFunc("/api/signup", signupHandler).Methods("POST")
	r.HandleFunc("/api/login", loginHandler).Methods("POST")
	r.HandleFunc("/api/google-signup", googleSignupHandler).Methods("POST")
	r.HandleFunc("/api/health", healthHandler).Methods("GET")
	r.HandleFunc("/api/users", getAllUsersHandler).Methods("GET")
	r.HandleFunc("/api/transcribe", handleTranscribe).Methods("POST")
	r.HandleFunc("/api/transcribe-youtube", handleTranscribeYouTube).Methods("POST")
	r.HandleFunc("/api/notes", handleNotes).Methods("POST")
	r.HandleFunc("/api/notes", handleNotes).Methods("GET")
	r.HandleFunc("/api/generate", handleGenerate).Methods("POST")
	r.HandleFunc("/api/materials", handleMaterials).Methods("POST", "GET")
	r.HandleFunc("/api/generate-and-save", handleGenerateAndSave).Methods("POST")
	r.HandleFunc("/api/materials/{id}", getMaterialByID).Methods("GET")
	r.HandleFunc("/api/notes/{id}", getNoteByID).Methods("GET")

	// Применяем CORS middleware
	handler := corsMiddleware(r)

	fmt.Println("🚀 SpeakApper Backend запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// Обработчик регистрации
func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли пользователь
	exists, err := UserExists(req.Email)
	if err != nil {
		log.Printf("Error checking user existence: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(SignupResponse{
			Success: false,
			Message: "User already exists",
		})
		return
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Создаем пользователя
	user := &User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  string(hashedPassword),
	}

	// Сохраняем в базу данных
	if err := CreateUser(user); err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Генерируем JWT токен
	token, err := generateJWT(user)
	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SignupResponse{
		Success: true,
		Message: "User registered successfully",
		User:    user,
		Token:   token,
	})
}

// Обработчик входа
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Получаем пользователя из базы данных
	user, err := GetUserByEmail(req.Email)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid credentials",
		})
		return
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid credentials",
		})
		return
	}

	// Генерируем JWT токен
	token, err := generateJWT(user)
	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Login successful",
		"user":    user,
		"token":   token,
	})
}

// Обработчик регистрации через Google
func googleSignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GoogleSignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидируем Google токен
	googleUser, err := ValidateGoogleToken(req.Token)
	if err != nil {
		log.Printf("Error validating Google token: %v", err)
		http.Error(w, "Invalid Google token", http.StatusUnauthorized)
		return
	}

	// Проверяем, существует ли пользователь
	exists, err := UserExists(googleUser.Email)
	if err != nil {
		log.Printf("Error checking user existence: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var user *User
	if exists {
		// Пользователь существует, получаем его данные
		user, err = GetUserByEmail(googleUser.Email)
		if err != nil {
			log.Printf("Error getting existing user: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		// Создаем нового пользователя
		user = &User{
			FirstName: googleUser.GivenName,
			LastName:  googleUser.FamilyName,
			Email:     googleUser.Email,
			Password:  "", // Google пользователи не имеют пароля
		}

		if err := CreateUser(user); err != nil {
			log.Printf("Error creating Google user: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	// Генерируем JWT токен
	token, err := generateJWT(user)
	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Google authentication successful",
		"user":    user,
		"token":   token,
	})
}

// Обработчик проверки здоровья сервера
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"time":   time.Now(),
	})
}

// Обработчик получения всех пользователей
func getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	users, err := GetAllUsers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Ошибка при получении пользователей",
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"users":   users,
		"count":   len(users),
	})
}

// Генерация JWT токена
func generateJWT(user *User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 дней
	})

	return token.SignedString(jwtSecret)
}
