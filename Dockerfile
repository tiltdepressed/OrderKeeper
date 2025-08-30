# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o orderkeeper ./cmd

# Final stage
FROM alpine:3.20.0
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/orderkeeper .
COPY --from=builder /app/web ./web
COPY --from=builder /app/docs ./docs
EXPOSE 8080
CMD ["./orderkeeper"]
