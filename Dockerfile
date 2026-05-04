# Stage 1: Build Web
FROM node:22-alpine AS web-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm install --legacy-peer-deps
COPY web/ .
RUN npm run build

# Stage 2: Build Server
FROM golang:1.26.1-alpine AS server-builder
ARG VERSION=development
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
RUN apk add --no-cache git
WORKDIR /app/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -a -installsuffix cgo \
    -ldflags "-X github.com/mcay23/hound/internal.Version=$VERSION -X github.com/mcay23/hound/internal.Commit=$COMMIT -X github.com/mcay23/hound/internal.BuildTime=$BUILD_TIME" \
    -o main .

# build command
# docker build \
#   --build-arg VERSION=v1.0.0 \
#   --build-arg COMMIT=$(git rev-parse --short HEAD) \
#   --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
#   -t hound:v1.0.0 .

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
ENV APP_ENV=production
EXPOSE 2323

# Data and Config volumes should be mapped in docker-compose
CMD ["./main"]
