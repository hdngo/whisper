FROM golang:1.23-alpine

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main cmd/server/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=0 /app/main .
COPY .env .

EXPOSE 6262

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:6262/health || exit 1

CMD ["./main"]