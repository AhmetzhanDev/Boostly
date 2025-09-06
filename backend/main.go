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
	"net/url"
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
)

// getEnvOrFile returns the value of the env var `key`.
// If empty, and there is a companion var `key + "_FILE"`, it reads the value from that file path.
func getEnvOrFile(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if p := os.Getenv(key + "_FILE"); p != "" {
		if b, err := os.ReadFile(p); err == nil {
			return strings.TrimSpace(string(b))
		}
	}
	return ""
}

// handleTranscribeYouTube transcribes YouTube by URL using yt-dlp + Whisper
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
		JSONErrorWithDetails(w, http.StatusFailedDependency, "yt-dlp is required on server", "Install with: brew install yt-dlp (mac) or pipx install yt-dlp")
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
	var cookiesArgs []string
	cp := getEnvOrFile("YTDLP_COOKIES")
	if cp == "" {
		log.Println("INFO: YTDLP_COOKIES environment variable not set. Will try --cookies-from-browser as fallback.")
		// Fallback: try to extract cookies from browser automatically
		cookiesArgs = []string{"--cookies-from-browser", "chrome"}
	} else if _, err := os.Stat(cp); err == nil {
		// случай 1: YTDLP_COOKIES = путь к файлу
		cookiesArgs = []string{"--cookies", cp}
		log.Printf("SUCCESS: yt-dlp will use cookies file at: %s", cp)
	} else {
		// случай 2: YTDLP_COOKIES = содержимое
		tmp := "/tmp/yt-cookies.txt"
		if err := os.WriteFile(tmp, []byte(cp), 0o600); err != nil {
			log.Printf("ERROR: failed to write cookies from env to file: %v", err)
			// Fallback to browser cookies if file creation fails
			cookiesArgs = []string{"--cookies-from-browser", "chrome"}
		} else {
			cookiesArgs = []string{"--cookies", tmp}
			log.Printf("SUCCESS: yt-dlp will use cookies from env, written to: %s", tmp)
		}
	}

	// Log chosen cookies mode and output pattern for diagnostics
	cookiesMode := "none"
	if len(cookiesArgs) >= 2 && cookiesArgs[0] == "--cookies" {
		cookiesMode = "file:" + cookiesArgs[1]
	} else if len(cookiesArgs) >= 2 && cookiesArgs[0] == "--cookies-from-browser" {
		cookiesMode = "from-browser:" + cookiesArgs[1]
	}
	log.Printf("yt-dlp: cookiesMode=%s outPattern=%s base=%s", cookiesMode, outPattern, base)

	// Define a helper to try different player clients with cookies if available
	tryClient := func(clientName string, format string) ([]byte, error) {
		args := append([]string{
			"-R", "3",
			"--fragment-retries", "3",
			"--force-ipv4",
			"--geo-bypass",
			"--no-check-certificate",
			"--add-header", "Accept-Language: en-US,en;q=0.9,ru;q=0.8",
			"--referer", "https://www.youtube.com/",
			"--extractor-args", fmt.Sprintf("youtube:player_client=%s", clientName),
			"-f", format,
			"-x",
			"--audio-format", "mp3",
			"-o", outPattern,
		}, cookiesArgs...)
		args = append(args, body.URL)
		log.Printf("yt-dlp try-client=%s args=%v", clientName, args)
		return ytdlpOutput(args...)
	}

	// Helper to check if error is authentication-related
	isAuthError := func(output []byte) bool {
		// Normalize quotes/case to catch messages like “you’re” vs "you're"
		s := strings.ToLower(string(output))
		s = strings.ReplaceAll(s, "’", "'")
		return strings.Contains(s, "sign in to confirm you're not a bot") ||
			strings.Contains(s, "sign in to confirm you") ||
			strings.Contains(s, "this video is not available") ||
			strings.Contains(s, "private video") ||
			strings.Contains(s, "video unavailable") ||
			strings.Contains(s, "cookies are no longer valid") ||
			strings.Contains(s, "use --cookies-from-browser") ||
			strings.Contains(s, "requires authentication")
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
						// attempt 6: try with updated yt-dlp and additional bypass options
						args3c := []string{
							"--cookies-from-browser", "chrome",
							"-R", "3",
							"--fragment-retries", "3",
							"--force-ipv4",
							"--geo-bypass",
							"--no-check-certificate",
							"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
							"--extractor-args", "youtube:player_client=web,youtube:skip=hls",
							"-f", "bestaudio[ext=m4a]/bestaudio/best",
							"-x",
							"--audio-format", "mp3",
							"-o", outPattern,
							body.URL,
						}
						log.Printf("yt-dlp retry (with Chrome cookies): args=%v", args3c)
						outBytes3c, err3c := ytdlpOutput(args3c...)
						if err3c != nil {
							log.Printf("yt-dlp chrome-cookies failed: %v; output: %s", err3c, string(outBytes3c))
							// Fallback 7: Try anonymous Piped API proxy (no YouTube auth required)
							pipedBase := getEnvOrFile("PIPED_INSTANCE")
							if pipedBase == "" {
								pipedBase = "https://piped.video"
							}
							// Extract video ID from URL
							var vid string
							if u, perr := url.Parse(body.URL); perr == nil {
								host := strings.ToLower(u.Host)
								path := strings.Trim(u.Path, "/")
								if strings.Contains(host, "youtu.be") {
									vid = path
								} else {
									qv := u.Query().Get("v")
									if qv != "" {
										vid = qv
									} else {
										// handle /shorts/{id}, /embed/{id}
										parts := strings.Split(path, "/")
										if len(parts) > 0 {
											vid = parts[len(parts)-1]
										}
									}
								}
							}
							if vid != "" {
								apiURL := fmt.Sprintf("%s/api/v1/streams/%s", strings.TrimRight(pipedBase, "/"), vid)
								log.Printf("Piped fallback: GET %s", apiURL)
								httpClient := &http.Client{Timeout: 60 * time.Second}
								resp, perr := httpClient.Get(apiURL)
								if perr == nil && resp != nil && resp.StatusCode == http.StatusOK {
									var piped struct {
										AudioStreams []struct {
											URL      string `json:"url"`
											Bitrate  int    `json:"bitrate"`
											MimeType string `json:"mimeType"`
										} `json:"audioStreams"`
									}
									bodyBytes, _ := io.ReadAll(resp.Body)
									resp.Body.Close()
									if jerr := json.Unmarshal(bodyBytes, &piped); jerr == nil && len(piped.AudioStreams) > 0 {
										// pick highest bitrate
										best := piped.AudioStreams[0]
										for _, s := range piped.AudioStreams[1:] {
											if s.Bitrate > best.Bitrate {
												best = s
											}
										}
										// decide extension by mime
										ext := ".m4a"
										if strings.Contains(strings.ToLower(best.MimeType), "webm") {
											ext = ".webm"
										}
										// download stream
										outAlt := filepath.Join(tmpDir, base+ext)
										log.Printf("Piped fallback: downloading %s -> %s", best.URL, outAlt)
										sresp, gerr := httpClient.Get(best.URL)
										if gerr == nil && sresp != nil && sresp.StatusCode == http.StatusOK {
											f, ferr := os.Create(outAlt)
											if ferr == nil {
												_, cerr := io.Copy(f, sresp.Body)
												f.Close()
												sresp.Body.Close()
												if cerr == nil {
													// Success: proceed with this file
													log.Printf("Piped fallback: saved %s", outAlt)
												} else {
													log.Printf("Piped fallback: write error: %v", cerr)
												}
											} else {
												log.Printf("Piped fallback: create file error: %v", ferr)
											}
										} else if sresp != nil {
											if sresp.Body != nil {
												sresp.Body.Close()
											}
											log.Printf("Piped fallback: stream GET failed: %s", sresp.Status)
										} else if gerr != nil {
											log.Printf("Piped fallback: stream GET error: %v", gerr)
										}
									} else {
										if jerr != nil {
											log.Printf("Piped fallback: JSON parse error: %v; body=%s", jerr, string(bodyBytes))
										} else {
											log.Printf("Piped fallback: no audioStreams found")
										}
									}
								} else {
									if resp != nil {
										io.Copy(io.Discard, resp.Body)
										resp.Body.Close()
										log.Printf("Piped fallback: API status=%v", resp.Status)
									}
									if perr != nil {
										log.Printf("Piped fallback: API error: %v", perr)
									}
								}
							} else {
								log.Printf("Piped fallback: cannot extract video ID from URL: %s", body.URL)
							}
							// After Piped attempt, continue to file detection below (base.*)
						}
					}
				}
			}
		}
	}

	// Collect all yt-dlp logs for debugging
	var ytdlpLogs []string
	var allOutputs [][]byte
	if err != nil {
		ytdlpLogs = append(ytdlpLogs, fmt.Sprintf("yt-dlp attempt1 (web) failed: %v; output: %s", err, string(outBytes)))
		allOutputs = append(allOutputs, outBytes)
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
			// Collect all previous logs
			// This is a bit manual, but will capture the flow for debugging
			// The `err` variable holds the last error in the chain
			finalErrorDetails := strings.Join(ytdlpLogs, "\n")

			// Check if this is an authentication error from any attempt
			isAuth := false
			for _, output := range allOutputs {
				if isAuthError(output) {
					isAuth = true
					break
				}
			}
			if isAuth {
				JSONErrorWithDetails(w, http.StatusForbidden,
					"YouTube requires authentication. Please ensure cookies are properly configured.",
					"This video requires sign-in to access. The server needs valid YouTube cookies to download age-restricted or private content.\n\nDetails:\n"+finalErrorDetails)
			} else {
				JSONErrorWithDetails(w, http.StatusInternalServerError, "Audio file not found after download", finalErrorDetails)
			}
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
	// Convert "auto" to empty string for OpenAI Whisper API
	language := body.Language
	if language == "auto" {
		language = ""
	}
	text, err := transcribeLongAudio(outPath, language)
	if err != nil {
		log.Printf("YouTube segmented transcription error: %v", err)
		JSONErrorWithDetails(w, http.StatusInternalServerError, "Transcription failed", err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, map[string]interface{}{
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
		JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Debug: вход в обработчик
	log.Println("[handleGenerateAndSave] start")

	// Извлекаем пользователя из JWT
	auth := extractUserFromJWT(w, r)
	if auth == nil {
		return
	}
	userID := auth.UserID
	log.Printf("[handleGenerateAndSave] userID=%s", userID.Hex())

	// Читаем тело запроса
	var reqBody GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("[handleGenerateAndSave] decode body error: %v", err)
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if reqBody.Transcript == "" {
		log.Println("[handleGenerateAndSave] empty transcript")
		JSONError(w, http.StatusBadRequest, "Transcript is required")
		return
	}

	// Подсчёт желаемого числа вопросов от объёма текста
	words := len(strings.Fields(reqBody.Transcript))
	targetQuiz := 0
	if words > 0 {
		targetQuiz = words / 90 // ~1 вопрос на 120 слов
		if targetQuiz < 8 {
			targetQuiz = 8
		}
		if targetQuiz > 30 {
			targetQuiz = 30
		}
	}
	log.Printf("[handleGenerateAndSave] transcript_len=%d words=%d targetQuiz~=%d", len(reqBody.Transcript), words, targetQuiz)

	// Улучшенный промпт для флешкарточек, квиза и краткого summary с подсчетом слов
	systemPrompt := `Ты — профессиональный генератор flashcards, quiz и кратких конспектов (summary) для эффективного обучения.
Я дам тебе текст. Твоя задача — сначала посчитать количество слов, а затем создать оптимальное количество карточек и вопросов по правилам ниже.

1. Подсчёт слов
Сначала посчитай количество слов в предоставленном тексте.

2. Определение количества карточек
Количество карточек выбирается автоматически в зависимости от объёма текста:
≤ 500 слов: 8–12 карточек (Базовые факты)
500–1500 слов: 15–25 карточек (Ключевые идеи)  
1500–3000 слов: 25–40 карточек (Термины + концепции)
> 3000 слов: 40–60 карточек макс. (Глубокое понимание)

3. Разделение по уровням сложности (только если слов больше 1500)
Базовый уровень → термины, определения, даты.
Средний уровень → ключевые идеи, факты, основные события.
Продвинутый уровень → глубокие взаимосвязи, анализ, выводы.

4. Важные правила
Создавай только те карточки, для которых есть информация в тексте.
Не выдумывай факты и не придумывай термины.
Старайся формулировать вопросы коротко, а ответы — ёмко.
Если текста мало — карточек будет меньше.
Если текста много — карточек будет больше, но не делай их перегруженными.
ИСПОЛЬЗУЙ ЯЗЫК ИСХОДНОГО ТЕКСТА: на каком языке дан текст, на том языке и создавай карточки.

5. Quiz правила
Ты — генератор комплексных квизов.  
На вход даётся учебный текст.  
Создай набор вопросов, состоящий из двух частей:
1. Вопросы с вариантами ответов (4 опции, один правильный).
2. Вопросы формата True/False.

Правила:
1. Количество вопросов зависит от объёма текста:
   • до 500 слов → минимум 8 вопросов  
   • 500–1000 слов → 10–12 вопросов  
   • 1000–1500 слов → 15–18 вопросов  
   • больше 1500 слов → 22–30 вопросов.
2. Разделяй блоки: сначала идут вопросы с вариантами ответов, потом True/False.
3. ИСПОЛЬЗУЙ ЯЗЫК ИСХОДНОГО ТЕКСТА: на каком языке дан текст, на том языке и создавай вопросы.
4. Формат вывода:
{
  "multipleChoice": [
    {
      "question": "Вопрос...",
      "options": ["A", "B", "C", "D"],
      "answer": "Правильный вариант"
    }
  ],
  "trueFalse": [
    {
      "statement": "Утверждение...",
      "answer": true
    }
  ]
}

6. Summary (краткий конспект)
Сгенерируй краткий, структурированный summary по тексту (на языке исходного текста), объёмом ~120–180 слов ИЛИ 5–7 сжатых пунктов. Фокус на ключевых идеях, фактах, определениях и выводах. Без воды, без выдумок.
Разрешён формат Markdown (включая списки, жирный/курсив, заголовки, ССЫЛКИ И ТАБЛИЦЫ). Если уместно, можешь включить небольшую Markdown-таблицу для сравнения или структурирования данных.

ФОРМАТ ОТВЕТА: Верни только JSON с полями: { "flashcards": [...], "quiz": ..., "summary": "...", "languageCode"? }. Каждая flashcard должна содержать {term, definition, example?}. Каждый quiz вопрос должен содержать {id?, type, question, options?, answer?, correct?, rationale?}.`

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
		JSONErrorWithDetails(w, http.StatusInternalServerError, "Failed to generate materials", err.Error())
		return
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to read response")
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("[handleGenerateAndSave] OpenAI chat API error: %s - %s", resp.Status, string(respBytes))
		JSONErrorWithDetails(w, http.StatusInternalServerError, "Generation failed", resp.Status)
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
		JSONError(w, http.StatusInternalServerError, "Failed to parse OpenAI response")
		return
	}
	if len(openaiResp.Choices) == 0 {
		JSONError(w, http.StatusInternalServerError, "Empty OpenAI response")
		return
	}

	// Parse raw response first
	var payloadRaw GeneratePayloadRaw
	var payload GeneratePayload
	parsed := false

	if err := json.Unmarshal([]byte(openaiResp.Choices[0].Message.Content), &payloadRaw); err != nil {
		clean := strings.TrimSpace(openaiResp.Choices[0].Message.Content)
		clean = strings.TrimPrefix(clean, "```json")
		clean = strings.TrimPrefix(clean, "```")
		clean = strings.TrimSuffix(clean, "```")
		clean = strings.TrimSpace(clean)
		if err2 := json.Unmarshal([]byte(clean), &payloadRaw); err2 != nil {
			// Попытка 3: извлечь JSON по первым и последним фигурным скобкам
			raw := openaiResp.Choices[0].Message.Content
			i := strings.Index(raw, "{")
			j := strings.LastIndex(raw, "}")
			if i >= 0 && j > i {
				candidate := strings.TrimSpace(raw[i : j+1])
				// Нормализация известных нестандартных ключей от модели
				candidate = strings.ReplaceAll(candidate, "\"example?\"", "\"example\"")
				if err3 := json.Unmarshal([]byte(candidate), &payloadRaw); err3 == nil {
					log.Printf("[handleGenerateAndSave] Parsed AI JSON via brace-extract normalization")
					parsed = true
				} else {
					log.Printf("[handleGenerateAndSave] Unmarshal failed after brace-extract: %v", err3)
				}
			}
			if !parsed {
				// Не валидный JSON: фиксируем и переходим к бэкенд-фоллбеку ниже
				log.Printf("[handleGenerateAndSave] Failed to unmarshal AI JSON: %v; content: %s", err, openaiResp.Choices[0].Message.Content)
			}
		} else {
			parsed = true
		}
	} else {
		parsed = true
	}

	if parsed {
		// Convert to final payload
		payload = GeneratePayload{
			Flashcards:   payloadRaw.Flashcards,
			LanguageCode: payloadRaw.LanguageCode,
			Summary:      payloadRaw.Summary,
		}

		// Parse quiz structure
		if len(payloadRaw.Quiz) > 0 {
			// Try parsing as array first (old format)
			var quizArray []QuizQuestion
			if err := json.Unmarshal(payloadRaw.Quiz, &quizArray); err == nil {
				payload.Quiz = quizArray
			} else {
				// Try parsing as structured format (new format)
				var aiQuiz AIQuizStructure
				if err := json.Unmarshal(payloadRaw.Quiz, &aiQuiz); err == nil {
					// Convert to QuizQuestion array
					var convertedQuiz []QuizQuestion

					// Add MCQ questions
					for i, mcq := range aiQuiz.MultipleChoice {
						convertedQuiz = append(convertedQuiz, QuizQuestion{
							ID:       FlexString(fmt.Sprintf("%d", i+1)),
							Type:     "MCQ",
							Question: mcq.Question,
							Options:  mcq.Options,
							Answer:   mcq.Answer,
							Correct:  true,
						})
					}

					// Add TF questions
					for i, tf := range aiQuiz.TrueFalse {
						answer := "False"
						if tf.Answer {
							answer = "True"
						}
						convertedQuiz = append(convertedQuiz, QuizQuestion{
							ID:       FlexString(fmt.Sprintf("%d", len(aiQuiz.MultipleChoice)+i+1)),
							Type:     "TF",
							Question: tf.Statement,
							Options:  []string{"True", "False"},
							Answer:   answer,
							Correct:  tf.Answer,
						})
					}

					payload.Quiz = convertedQuiz
					log.Printf("[handleGenerateAndSave] Converted AI quiz: MCQ=%d TF=%d total=%d", len(aiQuiz.MultipleChoice), len(aiQuiz.TrueFalse), len(convertedQuiz))
				} else {
					log.Printf("[handleGenerateAndSave] Failed to parse quiz structure: %v", err)
					payload.Quiz = []QuizQuestion{}
				}
			}
		}
	}

	// Ensure non-nil slices
	if payload.Flashcards == nil {
		payload.Flashcards = []Flashcard{}
	}
	if payload.Quiz == nil {
		payload.Quiz = []QuizQuestion{}
	}

	// Фоллбек: если модель вернула пустые массивы — синтезируем минимум 3 карточки и 3 вопроса из транскрипта
	if len(payload.Flashcards) == 0 && len(payload.Quiz) == 0 {
		text := strings.TrimSpace(reqBody.Transcript)
		low := strings.ToLower(text)
		wordsArr := strings.Fields(low)
		trimPunct := func(s string) string { return strings.Trim(s, ".,!?:;\"'()[]{}<>«»—-") }
		stop := map[string]struct{}{
			"и": {}, "в": {}, "во": {}, "на": {}, "что": {}, "это": {}, "как": {}, "к": {}, "из": {}, "по": {},
			"а": {}, "но": {}, "ли": {}, "да": {}, "не": {}, "ни": {}, "для": {}, "о": {}, "от": {}, "до": {},
			"the": {}, "a": {}, "an": {}, "to": {}, "of": {}, "in": {}, "on": {}, "is": {}, "are": {}, "with": {},
		}
		freq := map[string]int{}
		for _, w := range wordsArr {
			w = trimPunct(w)
			if len(w) < 3 {
				continue
			}
			if _, ok := stop[w]; ok {
				continue
			}
			freq[w]++
		}
		// выбрать до 3 наиболее частых слов без сортировки
		pickMax := func(m map[string]int) string {
			best := ""
			bestN := 0
			for k, n := range m {
				if n > bestN || (n == bestN && len(k) > len(best)) {
					best, bestN = k, n
				}
			}
			if best != "" {
				delete(m, best)
			}
			return best
		}
		termsSel := []string{}
		for i := 0; i < 3; i++ {
			w := pickMax(freq)
			if w == "" {
				break
			}
			termsSel = append(termsSel, w)
		}
		if len(termsSel) == 0 {
			termsSel = []string{"основы", "тема", "ключевой пункт"}
		}
		// флэшкарты
		fallbackCards := make([]Flashcard, 0, 3)
		for _, t := range termsSel {
			fallbackCards = append(fallbackCards, Flashcard{
				Term:       t,
				Definition: "Ключевой термин из транскрипта; уточните детали по контексту.",
				Example:    fmt.Sprintf("В тексте упоминается ‘%s’.", t),
			})
			if len(fallbackCards) >= 3 {
				break
			}
		}
		// вопросы True/False
		fallbackQuiz := make([]QuizQuestion, 0, 3)
		for _, t := range termsSel {
			q := QuizQuestion{
				Type:       "TF",
				Question:   fmt.Sprintf("В транскрипте упоминается ‘%s’?", t),
				Options:    []string{"True", "False"},
				Answer:     "True",
				Difficulty: "easy",
			}
			fallbackQuiz = append(fallbackQuiz, q)
			if len(fallbackQuiz) >= 3 {
				break
			}
		}
		if len(fallbackQuiz) == 0 {
			fallbackQuiz = []QuizQuestion{
				{Type: "TF", Question: "Транскрипт содержит приветствие?", Options: []string{"True", "False"}, Answer: "True", Difficulty: "easy"},
				{Type: "TF", Question: "Упоминается конкретная тема?", Options: []string{"True", "False"}, Answer: "True", Difficulty: "easy"},
				{Type: "TF", Question: "Это короткий фрагмент без подробностей?", Options: []string{"True", "False"}, Answer: "True", Difficulty: "easy"},
			}
		}
		payload.Flashcards = fallbackCards
		payload.Quiz = fallbackQuiz
		log.Printf("[handleGenerateAndSave] fallback used: flashcards=%d quiz=%d", len(payload.Flashcards), len(payload.Quiz))
	}

	// Diagnostics: log generation counts
	log.Printf("[handleGenerateAndSave] generated: flashcards=%d quiz=%d", len(payload.Flashcards), len(payload.Quiz))

	// Сохраняем материал в MongoDB с привязкой к пользователю
	material := Material{
		UserID:     userID,
		Transcript: reqBody.Transcript,
		Summary:    payload.Summary,
		Flashcards: payload.Flashcards,
		Quiz:       payload.Quiz,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	log.Printf("[handleGenerateAndSave] inserting material: user=%s flashcards=%d quiz=%d", userID.Hex(), len(payload.Flashcards), len(payload.Quiz))
	collection := client.Database("speakapper").Collection("materials")
	ctxIns, cancelIns := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelIns()
	startIns := time.Now()
	result, err := collection.InsertOne(ctxIns, material)
	if err != nil {
		log.Printf("[handleGenerateAndSave] Error saving material: %v", err)
		JSONErrorWithDetails(w, http.StatusInternalServerError, "Failed to save material", err.Error())
		return
	}
	log.Printf("[handleGenerateAndSave] inserted material _id=%v (type=%T) in %s", result.InsertedID, result.InsertedID, time.Since(startIns))
	material.ID = result.InsertedID.(primitive.ObjectID)

	// Guard against null slices in JSON
	respFlash := material.Flashcards
	if respFlash == nil {
		respFlash = []Flashcard{}
	}
	respQuiz := material.Quiz
	if respQuiz == nil {
		respQuiz = []QuizQuestion{}
	}
	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"material":   material,
		"flashcards": respFlash,
		"quiz":       respQuiz,
		"summary":    material.Summary,
	})
}

// LoginRequest представляет запрос на вход

// JWT секретный ключ (в продакшене используйте переменную окружения)
var jwtSecret []byte

// OpenAI API ключ теперь берём из переменной окружения
var openaiAPIKey string

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
			f := mat.Flashcards
			if f == nil {
				f = []Flashcard{}
			}
			q := mat.Quiz
			if q == nil {
				q = []QuizQuestion{}
			}
			responseMaterials = append(responseMaterials, map[string]interface{}{
				"id":         mat.ID.Hex(),
				"transcript": mat.Transcript,
				"flashcards": f,
				"quiz":       q,
				"created_at": mat.CreatedAt,
				"updated_at": mat.UpdatedAt,
			})
		}

		JSONResponse(w, http.StatusOK, map[string]interface{}{
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

	// Default nil to empty slices
	if materialData.Flashcards == nil {
		materialData.Flashcards = []Flashcard{}
	}
	if materialData.Quiz == nil {
		materialData.Quiz = []QuizQuestion{}
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
	log.Printf("[handleMaterials:POST] inserting material: user=%s flashcards=%d quiz=%d", userID.Hex(), len(material.Flashcards), len(material.Quiz))
	collection := client.Database("speakapper").Collection("materials")
	ctxIns, cancelIns := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelIns()
	startIns := time.Now()
	result, err := collection.InsertOne(ctxIns, material)
	if err != nil {
		log.Printf("[handleMaterials:POST] Error saving material: %v", err)
		http.Error(w, "Failed to save material", http.StatusInternalServerError)
		return
	}
	log.Printf("[handleMaterials:POST] inserted material _id=%v (type=%T) in %s", result.InsertedID, result.InsertedID, time.Since(startIns))

	// Set the ID from the inserted document
	material.ID = result.InsertedID.(primitive.ObjectID)

	// Return the created material
	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"material": material,
	})
}

func handleNotes(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token
	auth := extractUserFromJWT(w, r)
	if auth == nil {
		return
	}
	userID := auth.UserID

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

		JSONResponse(w, http.StatusOK, map[string]interface{}{
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
	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"note":    note,
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
	systemPrompt := `Ты — профессиональный генератор flashcards, quiz и кратких конспектов (summary) для эффективного обучения.
Я дам тебе текст. Твоя задача — сначала посчитать количество слов, а затем создать оптимальное количество карточек и вопросов по правилам ниже.

1. Подсчёт слов
Сначала посчитай количество слов в предоставленном тексте.

2. Определение количества карточек
Количество карточек выбирается автоматически в зависимости от объёма текста:
≤ 500 слов: 8–12 карточек (Базовые факты)
500–1500 слов: 15–25 карточек (Ключевые идеи)  
1500–3000 слов: 25–40 карточек (Термины + концепции)
> 3000 слов: 40–60 карточек макс. (Глубокое понимание)

3. Разделение по уровням сложности (только если слов больше 1500)
Базовый уровень → термины, определения, даты.
Средний уровень → ключевые идеи, факты, основные события.
Продвинутый уровень → глубокие взаимосвязи, анализ, выводы.

4. Важные правила
Создавай только те карточки, для которых есть информация в тексте.
Не выдумывай факты и не придумывай термины.
Старайся формулировать вопросы коротко, а ответы — ёмко.
Если текста мало — карточек будет меньше.
Если текста много — карточек будет больше, но не делай их перегруженными.
ИСПОЛЬЗУЙ ЯЗЫК ИСХОДНОГО ТЕКСТА: на каком языке дан текст, на том языке и создавай карточки.

5. Quiz правила
Ты — генератор комплексных квизов.  
На вход даётся учебный текст.  
Создай набор вопросов, состоящий из двух частей:
1. Вопросы с вариантами ответов (4 опции, один правильный).
2. Вопросы формата True/False.

Правила:
1. Количество вопросов зависит от объёма текста:
   • до 500 слов → минимум 8 вопросов  
   • 500–1000 слов → 10–12 вопросов  
   • 1000–1500 слов → 15–18 вопросов  
   • больше 1500 слов → 22–30 вопросов.
2. Разделяй блоки: сначала идут вопросы с вариантами ответов, потом True/False.
3. ИСПОЛЬЗУЙ ЯЗЫК ИСХОДНОГО ТЕКСТА: на каком языке дан текст, на том языке и создавай вопросы.
4. Формат вывода:
{
  "multipleChoice": [
    {
      "question": "Вопрос...",
      "options": ["A", "B", "C", "D"],
      "answer": "Правильный вариант"
    }
  ],
  "trueFalse": [
    {
      "statement": "Утверждение...",
      "answer": true
    }
  ]
}

6. Summary (краткий конспект)
Сгенерируй краткий, структурированный summary по тексту (на языке исходного текста), объёмом ~120–180 слов ИЛИ 5–7 сжатых пунктов. Фокус на ключевых идеях, фактах, определениях и выводах. Без воды, без выдумок.

ФОРМАТ ОТВЕТА: Верни только JSON с полями: { "flashcards": [...], "quiz": ..., "summary": "...", "languageCode"? }. Каждая flashcard должна содержать {term, definition, example?}. Каждый quiz вопрос должен содержать {id?, type, question, options?, answer?, correct?, rationale?}.`

	chatReq := map[string]interface{}{
		"model":           "gpt-4o-mini",
		"temperature":     0.3,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
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

	// Parse raw response first
	var payloadRaw GeneratePayloadRaw
	if err := json.Unmarshal([]byte(openaiResp.Choices[0].Message.Content), &payloadRaw); err != nil {
		// If model returned fenced code block, try strip
		clean := openaiResp.Choices[0].Message.Content
		clean = strings.TrimSpace(clean)
		// naive cleanup of Markdown fences
		clean = strings.TrimPrefix(clean, "```json")
		clean = strings.TrimPrefix(clean, "```")
		clean = strings.TrimSuffix(clean, "```")
		if err2 := json.Unmarshal([]byte(clean), &payloadRaw); err2 != nil {
			log.Printf("Failed to unmarshal AI JSON: %v; content: %s", err, openaiResp.Choices[0].Message.Content)
			http.Error(w, "Invalid JSON from model", http.StatusInternalServerError)
			return
		}
	}

	// Convert to final payload
	payload := GeneratePayload{
		Flashcards:   payloadRaw.Flashcards,
		LanguageCode: payloadRaw.LanguageCode,
		Summary:      payloadRaw.Summary,
	}

	// Parse quiz structure
	if len(payloadRaw.Quiz) > 0 {
		// Try parsing as array first (old format)
		var quizArray []QuizQuestion
		if err := json.Unmarshal(payloadRaw.Quiz, &quizArray); err == nil {
			payload.Quiz = quizArray
		} else {
			// Try parsing as structured format (new format)
			var aiQuiz AIQuizStructure
			if err := json.Unmarshal(payloadRaw.Quiz, &aiQuiz); err == nil {
				// Convert to QuizQuestion array
				var convertedQuiz []QuizQuestion

				// Add MCQ questions
				for i, mcq := range aiQuiz.MultipleChoice {
					convertedQuiz = append(convertedQuiz, QuizQuestion{
						ID:       FlexString(fmt.Sprintf("%d", i+1)),
						Type:     "MCQ",
						Question: mcq.Question,
						Options:  mcq.Options,
						Answer:   mcq.Answer,
						Correct:  true,
					})
				}

				// Add TF questions
				for i, tf := range aiQuiz.TrueFalse {
					answer := "False"
					if tf.Answer {
						answer = "True"
					}
					convertedQuiz = append(convertedQuiz, QuizQuestion{
						ID:       FlexString(fmt.Sprintf("%d", len(aiQuiz.MultipleChoice)+i+1)),
						Type:     "TF",
						Question: tf.Statement,
						Options:  []string{"True", "False"},
						Answer:   answer,
						Correct:  tf.Answer,
					})
				}

				payload.Quiz = convertedQuiz
				log.Printf("[handleGenerate] Converted AI quiz: MCQ=%d TF=%d total=%d", len(aiQuiz.MultipleChoice), len(aiQuiz.TrueFalse), len(convertedQuiz))
			} else {
				log.Printf("[handleGenerate] Failed to parse quiz structure: %v", err)
				payload.Quiz = []QuizQuestion{}
			}
		}
	}

	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"flashcards": payload.Flashcards,
		"quiz":       payload.Quiz,
		"summary":    payload.Summary,
	})
}

// Load environment variables from .env-like files without external deps.
// Only sets a key if it's not already present in the process environment.
func loadDotEnv(paths ...string) {
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			continue
		}
		lines := strings.Split(string(b), "\n")
		for _, line := range lines {
			s := strings.TrimSpace(line)
			if s == "" || strings.HasPrefix(s, "#") { // skip comments/empty
				continue
			}
			// allow export PREFIX
			if strings.HasPrefix(s, "export ") {
				s = strings.TrimSpace(strings.TrimPrefix(s, "export "))
			}
			// split by first '='
			eq := strings.Index(s, "=")
			if eq <= 0 {
				continue
			}
			key := strings.TrimSpace(s[:eq])
			val := strings.TrimSpace(s[eq+1:])
			// strip surrounding quotes
			if len(val) >= 2 {
				if (val[0] == '\'' && val[len(val)-1] == '\'') || (val[0] == '"' && val[len(val)-1] == '"') {
					val = val[1 : len(val)-1]
				}
			}
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
		log.Printf("🔧 Loaded env from %s", p)
	}
}

// spaHandler returns a handler that serves files from dist, falling back to index.html
func spaHandler(dist string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath := strings.TrimPrefix(r.URL.Path, "/")
		// Protect against path traversal
		reqPath = filepath.Clean(reqPath)
		filePath := filepath.Join(dist, reqPath)
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, filePath)
			return
		}

		// Fallback to index.html for SPA routes
		http.ServeFile(w, r, filepath.Join(dist, "index.html"))
	})
}

func main() {
	// Load .env files so `go run .` works without manual exports
	// Search project root and current dir when running from backend/
	loadDotEnv(filepath.Join("..", ".env"), ".env")

	// Читаем OpenAI API ключ из переменной окружения
	openaiAPIKey = getEnvOrFile("OPENAI_API_KEY")
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
	if database != nil {
		log.Printf("✅ Mongo database selected: %s", database.Name())
	}
	r := mux.NewRouter()

	// Настройка CORS
	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{
			"http://localhost:3001",
			"http://localhost:3000",
			"http://127.0.0.1:3001",
			"http://127.0.0.1:3000",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
		handlers.ExposedHeaders([]string{"Cross-Origin-Opener-Policy"}),
	)

	// Add COOP headers middleware
	coopMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set Cross-Origin-Opener-Policy to allow cross-origin communication
			w.Header().Set("Cross-Origin-Opener-Policy", "unsafe-none")
			w.Header().Set("Cross-Origin-Embedder-Policy", "unsafe-none")
			next.ServeHTTP(w, r)
		})
	}

	// Роуты
	r.HandleFunc("/api/signup", signupHandler).Methods("POST")
	r.HandleFunc("/api/login", loginHandler).Methods("POST")
	r.HandleFunc("/api/google-signup", googleSignupHandler).Methods("POST")
	r.HandleFunc("/api/health", healthHandler).Methods("GET")
	r.HandleFunc("/api/users", getAllUsersHandler).Methods("GET")
	r.HandleFunc("/api/user", getUserHandler).Methods("GET")
	r.HandleFunc("/api/transcribe", handleTranscribe).Methods("POST")
	r.HandleFunc("/api/transcribe-youtube", handleTranscribeYouTube).Methods("POST")
	r.HandleFunc("/api/notes", handleNotes).Methods("POST")
	r.HandleFunc("/api/notes", handleNotes).Methods("GET")
	r.HandleFunc("/api/generate", handleGenerate).Methods("POST")
	r.HandleFunc("/api/materials", handleMaterials).Methods("POST", "GET")
	r.HandleFunc("/api/generate-and-save", handleGenerateAndSave).Methods("POST")
	r.HandleFunc("/api/notes/{id}", getNoteByID).Methods("GET")
	r.HandleFunc("/api/notes/{id}", deleteNoteByID).Methods("DELETE")
	r.HandleFunc("/api/materials/{id}", getMaterialByID).Methods("GET")
	r.HandleFunc("/api/materials/{id}", deleteMaterialByID).Methods("DELETE")

	// Serve Vite build (dist) with SPA fallback
	distPath := os.Getenv("FRONTEND_DIST")
	if distPath == "" {
		// Support Cloud Run / buildpacks convention
		distPath = os.Getenv("STATIC_DIR")
	}
	if distPath != "" {
		log.Printf("📦 Serving frontend from: %s", distPath)
		// Use NotFoundHandler so API routes take precedence
		r.NotFoundHandler = spaHandler(distPath)
	} else {
		log.Printf("⚠️ FRONTEND_DIST not set and no dist/index.html found; SPA serving disabled")
	}

	// Применяем CORS и COOP middleware
	handler := corsMiddleware(coopMiddleware(r))

	// Read PORT from env for container platforms (default 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("HOST")
	// В контейнере Cloud Run всегда форсим 0.0.0.0
	if host == "" || host == "localhost" || host == "127.0.0.1" {
		host = "0.0.0.0"
	}

	addr := host + ":" + port
	fmt.Printf("🚀 SpeakApper Backend listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
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
