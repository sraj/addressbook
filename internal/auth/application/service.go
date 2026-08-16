package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sraj/addressbook/internal/auth/domain"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	users     domain.Repository
	jwtSecret string
}

func NewService(users domain.Repository, jwtSecret string) *Service {
	return &Service{users: users, jwtSecret: jwtSecret}
}

func (s *Service) Register(ctx context.Context, email, password string) (*domain.User, error) {
	exists, err := s.users.Exists(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(hash),
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if user.Status == "suspended" {
		return nil, domain.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	return user, nil
}

func (s *Service) GenerateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": float64(userID),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *Service) UserByID(ctx context.Context, id uint) (*domain.User, error) {
	return s.users.FindByID(ctx, id)
}

func (s *Service) IsActive(ctx context.Context, userID uint) (bool, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}
	return user.Status != "suspended", nil
}

func (s *Service) ValidateToken(tokenStr string) (uint, error) {
	secret := s.jwtSecret
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, domain.ErrInvalidToken
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, domain.ErrInvalidToken
	}
	return uint(userIDFloat), nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID uint, name string, prefs *domain.Preferences) (*domain.User, error) {
	if err := s.users.UpdateName(ctx, userID, name); err != nil {
		return nil, err
	}
	if prefs != nil {
		data, _ := json.Marshal(prefs)
		if err := s.users.UpdatePreferences(ctx, userID, string(data)); err != nil {
			return nil, err
		}
	}
	return s.UserByID(ctx, userID)
}

func (s *Service) ForgotPassword(ctx context.Context, email string) (string, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", domain.ErrUserNotFound
		}
		return "", err
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	resetToken := &domain.PasswordResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
	}

	if err := s.users.CreateResetToken(ctx, resetToken); err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	resetToken, err := s.users.FindResetToken(ctx, token)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidToken) {
			return domain.ErrInvalidToken
		}
		return err
	}

	if resetToken.Used {
		return domain.ErrInvalidToken
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return domain.ErrInvalidToken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.users.UpdatePassword(ctx, resetToken.UserID, string(hash)); err != nil {
		return err
	}

	if err := s.users.MarkResetTokenUsed(ctx, resetToken.ID); err != nil {
		return err
	}

	return nil
}

func (s *Service) ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.ErrUserNotFound
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return domain.ErrInvalidCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.users.UpdatePassword(ctx, userID, string(hash))
}
