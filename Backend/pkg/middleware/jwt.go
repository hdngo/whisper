package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hdngo/whisper/internal/cache"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UsernameKey contextKey = "username"
)

type JWTMiddleware struct {
	jwtSecret   string
	redisClient *cache.RedisClient
}

func NewJWTMiddleware(jwtSecret string, redisClient *cache.RedisClient) *JWTMiddleware {
	return &JWTMiddleware{
		jwtSecret:   jwtSecret,
		redisClient: redisClient,
	}
}

func (m *JWTMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			http.Error(w, "invalid token format", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		userID := claims["user_id"].(float64)
		storedToken, err := m.redisClient.GetSession(r.Context(), int64(userID))
		if err != nil || storedToken != bearerToken[1] {
			http.Error(w, "session expired or invalid", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims["user_id"])
		ctx = context.WithValue(ctx, UsernameKey, claims["username"])

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
