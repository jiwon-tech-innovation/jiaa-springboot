# Go Multistage Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies (git for modules, build-base for CGO, librdkafka-dev for dynamic linking)
RUN apk add --no-cache git build-base librdkafka-dev pkgconf

# Copy Mod files
COPY go.mod go.sum ./

# Copy Source
COPY . .

# Tidy dependencies (to resolve go.mod updates)
RUN go mod tidy -e && go mod vendor

# Build Argument to select service (default: input-service)
ARG SERVICE_NAME=input-service
RUN go build -tags dynamic -o /server cmd/${SERVICE_NAME}/main.go

# Runtime Stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache librdkafka

WORKDIR /root/
COPY --from=builder /server .

# Expose HTTP Port
EXPOSE 8080

CMD ["./server"]
