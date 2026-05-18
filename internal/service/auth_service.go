package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yuwanadev/mdm-backend/internal/models"
	"github.com/yuwanadev/mdm-backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

type AuthService struct {
	userRepo  *repository.UserRepo
	jwtSecret []byte
}

func NewAuthService(userRepo *repository.UserRepo, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

// TokenPair holds access and refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}

// Login authenticates a user and returns JWT tokens.
func (s *AuthService) Login(ctx context.Context, username, password string) (*TokenPair, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokens(user)
}

// RefreshAccessToken generates a new access token from a valid refresh token.
func (s *AuthService) RefreshAccessToken(refreshTokenStr string) (*TokenPair, error) {
	token, err := jwt.Parse(refreshTokenStr, func(t *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidCredentials
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidCredentials
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return nil, ErrInvalidCredentials
	}

	userIDStr, _ := claims["sub"].(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByID(context.Background(), userID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokens(user)
}

// SeedAdmin creates the admin user if none exists.
func (s *AuthService) SeedAdmin(ctx context.Context, username, password string) error {
	exists, err := s.userRepo.Exists(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil // admin already exists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.userRepo.Create(ctx, username, string(hash))
	return err
}

// RequiresSetup returns true if there are no users in the database.
func (s *AuthService) RequiresSetup(ctx context.Context) (bool, error) {
	exists, err := s.userRepo.Exists(ctx)
	if err != nil {
		return false, err
	}
	return !exists, nil
}

// ValidateToken parses and validates a JWT access token.
func (s *AuthService) ValidateToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, ErrInvalidCredentials
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, ErrInvalidCredentials
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "access" {
		return uuid.Nil, ErrInvalidCredentials
	}

	sub, _ := claims["sub"].(string)
	return uuid.Parse(sub)
}

func (s *AuthService) generateTokens(user *models.User) (*TokenPair, error) {
	accessExp := time.Now().Add(15 * time.Minute)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID.String(),
		"name": user.Username,
		"type": "access",
		"exp":  accessExp.Unix(),
		"iat":  time.Now().Unix(),
	})
	accessStr, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	refreshExp := time.Now().Add(7 * 24 * time.Hour)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID.String(),
		"type": "refresh",
		"exp":  refreshExp.Unix(),
		"iat":  time.Now().Unix(),
	})
	refreshStr, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresIn:    int64(time.Until(accessExp).Seconds()),
	}, nil
}

// HashPassword hashes a password using bcrypt.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
