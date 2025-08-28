# Use the official Golang image as the base image
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Use a smaller image for the final container
FROM alpine:3.18

# Install yt-dlp and ffmpeg (required for audio extraction)
RUN apk add --no-cache \
    python3 \
    py3-pip \
    ffmpeg \
    && pip3 install --no-cache-dir yt-dlp

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Make the binary executable
RUN chmod +x /main

# Expose the port the app runs on
EXPOSE 8081

# Command to run the application
CMD ["/main"]
