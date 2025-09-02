# ---------- Build stage ----------
    FROM golang:1.24-alpine AS builder
    # Если CGO не нужен — можно вообще без компиляторов
    # RUN apk add --no-cache gcc musl-dev
    WORKDIR /app
    
    COPY go.mod go.sum ./
    # Подстрахуемся, чтобы точно подтянулся нужный тулчейн 1.24.x
    ENV GOTOOLCHAIN=go1.24.4
    RUN go mod download
    
    COPY . .
    
    # Вариант без CGO (предпочтительно для Cloud Run: меньше образ, проще рантайм)
    RUN CGO_ENABLED=0 GOOS=linux go build -o main .
    # Если строго нужен CGO — раскомментируй и добавь apk gcc/musl-dev выше:
    # RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .
    
    # ---------- Runtime stage ----------
    FROM alpine:3.18
    
    RUN apk add --no-cache \
          ca-certificates \
          python3 \
          py3-pip \
          ffmpeg \
      && pip3 install --no-cache-dir yt-dlp
    
    WORKDIR /app
    COPY --from=builder /app/main /app/main
    COPY ./dist /app/dist
    
    # Безопасный рантайм-пользователь
    RUN adduser -D -g '' appuser && chown -R appuser:appuser /app
    USER appuser
    
    ENV STATIC_DIR=/app/dist
    # В образе можно оставить дефолт, но В Cloud Run НЕ передавай PORT через --set-env-vars
    ENV PORT=8080
    EXPOSE 8080
    
    CMD ["/app/main"]