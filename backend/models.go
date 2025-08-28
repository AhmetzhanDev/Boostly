package main

import (
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User представляет структуру пользователя
type User struct {
	ID        string    `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Не отправляем пароль в JSON
	CreatedAt time.Time `json:"createdAt"`
	// Поля подписки (Lemon Squeezy)
	Premium          bool      `bson:"premium,omitempty" json:"premium"`
	Plan             string    `bson:"plan,omitempty" json:"plan,omitempty"`
	LsSubscriptionID string    `bson:"ls_subscription_id,omitempty" json:"ls_subscription_id,omitempty"`
	CurrentPeriodEnd time.Time `bson:"current_period_end,omitempty" json:"current_period_end,omitempty"`
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
	ID         FlexString  `json:"id,omitempty"`
	Type       string      `json:"type,omitempty"` // MCQ, MSQ, CLOZE, TF, MATCHING, SHORT
	Question   string      `json:"question"`
	Options    []string    `json:"options,omitempty"`    // MCQ/MSQ/TF/CLOZE
	Answer     string      `json:"answer,omitempty"`     // Верный ответ для MCQ/TF/SHORT/CLOZE
	Correct    interface{} `json:"correct,omitempty"`    // Может быть bool (TF/MCQ) или []string (MSQ)
	Pairs      [][]string  `json:"pairs,omitempty"`      // MATCHING: массив пар [[left,right], ...]
	Rationale  string      `json:"rationale,omitempty"`  // Объяснение
	Difficulty string      `json:"difficulty,omitempty"` // easy|medium|hard
	Citation   string      `json:"citation,omitempty"`   // Цитата/ссылка на фрагмент транскрипта
}

// Промежуточная структура для парсинга quiz от AI
type AIQuizStructure struct {
	MultipleChoice []struct {
		Question string   `json:"question"`
		Options  []string `json:"options"`
		Answer   string   `json:"answer"`
	} `json:"multipleChoice"`
	TrueFalse []struct {
		Statement string `json:"statement"`
		Answer    bool   `json:"answer"`
	} `json:"trueFalse"`
}

// Структура для парсинга ответа от AI с гибким quiz полем
type GeneratePayloadRaw struct {
	Flashcards   []Flashcard     `json:"flashcards"`
	Quiz         json.RawMessage `json:"quiz"` // Сырой JSON для гибкого парсинга
	LanguageCode string          `json:"languageCode,omitempty"`
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
