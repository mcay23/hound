# Stage 1: Build Web
FROM node:22-alpine as web-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm install
COPY web/ .
RUN npm run build

# Stage 2: Build Server
FROM golang:1.24-alpine as server-builder
RUN apk add --no-cache git
WORKDIR /app/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Stage 3: Combined
FROM alpine:latest
RUN apk add --no-cache ffmpeg ca-certificates
WORKDIR /app

# Copy Go binary
COPY --from=server-builder /app/server/main .
# Copy React build
COPY --from=web-builder /app/web/build ./build

# Environment variables
ENV SERVER_PORT=2323
EXPOSE 2323

# Data and Config volumes should be mapped in docker-compose
CMD ["./main"]
