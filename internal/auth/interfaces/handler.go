package interfaces

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/mobentum/kern"
	"github.com/sraj/addressbook/internal/auth/application"
	"github.com/sraj/addressbook/internal/auth/domain"
	"github.com/sraj/addressbook/internal/mailer"
	"github.com/sraj/addressbook/internal/shared"
	"github.com/mobentum/kern/extensions/xvalidator"
)

func (h *Handler) trySend(to, subject, body string) {
	if h.mailer != nil {
		go func() {
			if err := h.mailer.Send(to, subject, body); err != nil {
				// Logged but not returned — email failures shouldn't block the request
				slog.Error("email send failed", "error", err)
			}
		}()
	}
}

type Handler struct {
	svc          authService
	billing      billingIniter
	mailer       mailSender
	appURL       string
	secureCookie bool
}

func NewHandler(svc authService, billing billingIniter, mailer mailSender, appURL string, secureCookie bool) *Handler {
	return &Handler{svc: svc, billing: billing, mailer: mailer, appURL: appURL, secureCookie: secureCookie}
}

func (h *Handler) setTokenCookie(c *kern.Context, token string) {
	http.SetCookie(c.Response, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/api",
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int((24 * time.Hour).Seconds()),
	})
}

func (h *Handler) clearTokenCookie(c *kern.Context) {
	http.SetCookie(c.Response, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/api",
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

func (h *Handler) Register(c *kern.Context) {
	req, ok := xvalidator.Validated[application.RegisterRequest](c.Context())
	if !ok {
		return
	}

	user, err := h.svc.Register(c.Context(), req.Email, req.Password)
	if err != nil {
		if err == domain.ErrEmailTaken {
			shared.SendError(c, http.StatusConflict, "email already registered")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "registration failed", err)
		return
	}

	if err := h.billing.InitAccount(c.Context(), user.ID, user.Email, user.Email); err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to initialize account", err)
		return
	}

	token, err := h.svc.GenerateToken(user.ID)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "token generation failed", err)
		return
	}

	h.setTokenCookie(c, token)
	h.trySend(req.Email, "Welcome to Address Book", mailer.WelcomeEmail(req.Email))
	_ = c.JSON(http.StatusCreated, application.AuthResponse{User: userToInfo(user)})
}

func (h *Handler) Login(c *kern.Context) {
	req, ok := xvalidator.Validated[application.LoginRequest](c.Context())
	if !ok {
		return
	}

	user, err := h.svc.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			shared.SendError(c, http.StatusUnauthorized, "invalid email or password")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "login failed", err)
		return
	}

	token, err := h.svc.GenerateToken(user.ID)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "token generation failed", err)
		return
	}

	h.setTokenCookie(c, token)
		_ = c.JSON(http.StatusOK, application.AuthResponse{User: userToInfo(user)})
}

func (h *Handler) Me(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		shared.SendError(c, http.StatusUnauthorized, "not authenticated")
		return
	}

	user, err := h.svc.UserByID(c.Context(), userID)
	if err != nil {
		shared.SendError(c, http.StatusUnauthorized, "user not found")
		return
	}

		_ = c.JSON(http.StatusOK, application.AuthResponse{User: userToInfo(user)})
}

func (h *Handler) Logout(c *kern.Context) {
	h.clearTokenCookie(c)
	c.NoContent(http.StatusNoContent)
}

func (h *Handler) ForgotPassword(c *kern.Context) {
	req, ok := xvalidator.Validated[application.ForgotPasswordRequest](c.Context())
	if !ok {
		return
	}

	token, err := h.svc.ForgotPassword(c.Context(), req.Email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			_ = c.JSON(http.StatusOK, map[string]string{"message": "If an account exists with that email, a reset link has been sent"})
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to process request", err)
		return
	}

	resetLink := h.appURL + "/reset-password?token=" + token
	h.trySend(req.Email, "Reset your password", mailer.ResetPasswordEmail(resetLink))
	_ = c.JSON(http.StatusOK, map[string]string{"message": "If an account exists with that email, a reset link has been sent"})
}

func (h *Handler) ResetPassword(c *kern.Context) {
	req, ok := xvalidator.Validated[application.ResetPasswordRequest](c.Context())
	if !ok {
		return
	}

	if err := h.svc.ResetPassword(c.Context(), req.Token, req.Password); err != nil {
		if err == domain.ErrInvalidToken {
			shared.SendError(c, http.StatusBadRequest, "invalid or expired reset token")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to reset password", err)
		return
	}

	_ = c.JSON(http.StatusOK, map[string]string{"message": "password reset successfully"})
}

func (h *Handler) ChangePassword(c *kern.Context) {
	req, ok := xvalidator.Validated[application.ChangePasswordRequest](c.Context())
	if !ok {
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	if err := h.svc.ChangePassword(c.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		if err == domain.ErrInvalidCredentials {
			shared.SendError(c, http.StatusUnauthorized, "current password is incorrect")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to change password", err)
		return
	}

	_ = c.JSON(http.StatusOK, map[string]string{"message": "password changed successfully"})
}

func (h *Handler) UpdateProfile(c *kern.Context) {
	req, ok := xvalidator.Validated[application.UpdateProfileRequest](c.Context())
	if !ok {
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	user, err := h.svc.UpdateProfile(c.Context(), userID, req.Name, req.Preferences)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to update profile", err)
		return
	}

	_ = c.JSON(http.StatusOK, application.AuthResponse{User: userToInfo(user)})
}

func userToInfo(u *domain.User) application.UserInfo {
	info := application.UserInfo{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
		Role:  u.Role,
	}
	if u.Preferences != "" {
		var prefs domain.Preferences
		if err := json.Unmarshal([]byte(u.Preferences), &prefs); err == nil {
			info.Preferences = &prefs
		}
	}
	return info
}
