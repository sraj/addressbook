package application

import "github.com/sraj/addressbook/internal/auth/domain"

type RegisterRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type AuthResponse struct {
	User UserInfo `json:"user"`
}

type UserInfo struct {
	ID          uint                `json:"id"`
	Email       string              `json:"email"`
	Name        string              `json:"name"`
	Role        string              `json:"role"`
	Preferences *domain.Preferences `json:"preferences,omitempty"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token"    validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password"     validate:"required,min=8"`
}

type UpdateProfileRequest struct {
	Name        string              `json:"name" validate:"required"`
	Preferences *domain.Preferences `json:"preferences,omitempty"`
}
