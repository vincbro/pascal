FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o pascal ./discord

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/pascal .

RUN touch .env

VOLUME ["/app/data"]

CMD ["./pascal"]
