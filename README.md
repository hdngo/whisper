# Whisper
A real-time web-based chat application built with Angular and Go.

![image](https://github.com/user-attachments/assets/6319c238-599c-4e75-b26a-575e0972bb31)
![image](https://github.com/user-attachments/assets/ef2bbe73-15d3-403e-b7ec-dc46a8dd617e)

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

2. Create a .env file based on .env.example
```bash
cp Backend\.env.example Backend\.env
```

3. Start the application using Docker Compose
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
Since frontend assumes the same port for API requests, it is recommended you have nginx setup to forward requests to their appropriate place.
If not, just set the API URL in `auth.service.ts` and WS URL in `websocket.service.ts` accordingly.
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
