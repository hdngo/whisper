# Whisper

A real-time web-based chat application built with Angular and Go.

## Features

- Real-time messaging using WebSocket
- JWT-based authentication
- Message persistence with PostgreSQL
- User presence indicators
- Session management with Redis
- Message history on room entry
- Timestamp display for messages

## Tech Stack

- **Frontend:** Angular 17, WebSocket
- **Backend:** Go 1.23
- **Databases:** PostgreSQL 16, Redis
- **DevOps:** Docker, Docker Compose, GitHub Actions
- **Testing:** Python (pytest)

## Prerequisites

- Docker and Docker Compose
- Node.js 20+ (for local development)
- Go 1.23+ (for local development)
- PostgreSQL 16+ (for local development)
- Redis (for local development)

## Quick Start

1. Clone the repository
```bash
git clone https://github.com/hdngo/whisper.git
cd whisper
```

2. Start the application using Docker Compose
```bash
docker-compose up -d
```

The application will be available at `http://localhost:80` (port can be changed via `docker-compose.yml`)

## Local Development

### Backend
```bash
cd Backend
cp .env.example .env  # Configure your environment variables
go mod download
go run cmd/server/main.go
```

### Frontend
```bash
cd Frontend
npm install
ng serve
```

## Testing

Run the test suite:
```bash
cd Backend/testing
python -m pip install -r requirements.txt
pytest
```

Run load tests:
```bash
python stress_test.py
```

## Project Structure

```
whisper/
├── Backend/         # Go backend service
├── Frontend/        # Angular frontend application
└── docker-compose.yml
```
