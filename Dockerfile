FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG BINARY
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/service ./cmd/${BINARY}

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/bin/service .

ENTRYPOINT ["/app/service"]