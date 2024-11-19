package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hdngo/whisper/internal/cache"
	"github.com/hdngo/whisper/internal/config"
	"github.com/hdngo/whisper/internal/handler"
	"github.com/hdngo/whisper/internal/repository"
	"github.com/hdngo/whisper/internal/service"
	"github.com/hdngo/whisper/internal/ws"
	"github.com/hdngo/whisper/pkg/middleware"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config", err)
	}

	// Initialize database
	db, err := initDB(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatal("Failed to run migrations", err)
	}

	// Initialize Redis
	redisClient, err := cache.NewRedisClient(cfg.RedisHost, cfg.RedisPort)
	if err != nil {
		log.Fatal("Failed to initialize Redis", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	msgRepo := repository.NewMessageRepository(db)

	// Initialize WebSocket hub
	hub := ws.NewHub(msgRepo)
	go hub.Run()

	// Initialize services
	authService := service.NewAuthService(userRepo, redisClient, cfg.JWTSecret)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	chatHandler := handler.NewChatHandler(hub, cfg.JWTSecret)
	messageHandler := handler.NewMessageHandler(msgRepo)

	// Initialize middleware
	jwtMiddleware := middleware.NewJWTMiddleware(cfg.JWTSecret, redisClient)

	// Setup router
	router := mux.NewRouter()

	// Add CORS middleware to all routes
	router.Use(middleware.CORSMiddleware)

	// Public routes
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/ws", chatHandler.HandleWebSocket)

	// Protected routes
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(jwtMiddleware.Authenticate)
	protected.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST", "OPTIONS")
	protected.HandleFunc("/messages/recent", messageHandler.GetRecent).Methods("GET", "OPTIONS")
	protected.HandleFunc("/messages/before/{id}", messageHandler.GetMessagesBefore).Methods("GET", "OPTIONS")

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server starting on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func initDB(cfg *config.Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password='%s' dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(100)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at BIGINT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users (username)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			content TEXT NOT NULL,
			user_id BIGINT NOT NULL,
			username VARCHAR(255) NOT NULL,
			created_at BIGINT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages (created_at)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %v", err)
		}
	}

	return nil
}
