package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hdngo/whisper/internal/cache"
	"github.com/hdngo/whisper/internal/model"
	"github.com/hdngo/whisper/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    *repository.UserRepository
	redisClient *cache.RedisClient
	jwtSecret   string
}

func NewAuthService(userRepo *repository.UserRepository, redisClient *cache.RedisClient, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		redisClient: redisClient,
		jwtSecret:   jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	if _, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, errors.New("username already exists")
	}

	if len(req.Username) < 4 {
		return nil, errors.New("username must be at least 4 characters long")
	}

	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters long")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:  req.Username,
		Password:  string(hashedPassword),
		CreatedAt: time.Now().Unix(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := s.generateToken(user)

	if err != nil {
		return nil, err
	}

	if err := s.redisClient.StoreSession(ctx, user.ID, token); err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token:    token,
		Username: user.Username,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := s.redisClient.DeleteSession(ctx, user.ID); err != nil {
		return nil, err
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	if err := s.redisClient.StoreSession(ctx, user.ID, token); err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token:    token,
		Username: user.Username,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID int64) error {
	return s.redisClient.DeleteSession(ctx, userID)
}

func (s *AuthService) generateToken(user *model.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(s.jwtSecret))
}
