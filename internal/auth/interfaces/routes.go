package interfaces

import (
	"github.com/mobentum/kern"
	"github.com/mobentum/kern/extensions/xvalidator"
	"github.com/sraj/addressbook/internal/auth/application"
)

func (h *Handler) RegisterRoutes(app *kern.App, jwtAuth kern.MiddlewareFunc, rateLimitAuth ...kern.MiddlewareFunc) {
	public := app.Group("/api/v1/auth", rateLimitAuth...)
	public.AddConstraints("POST", "/register", kern.Constraints{Validate: xvalidator.BodyValidator[application.
		RegisterRequest]()}, h.Register)
	public.AddConstraints("POST", "/login", kern.Constraints{Validate: xvalidator.BodyValidator[application.
		LoginRequest]()}, h.Login)
	public.AddConstraints("POST", "/forgot-password", kern.Constraints{Validate: xvalidator.BodyValidator[application.
		ForgotPasswordRequest]()}, h.ForgotPassword)
	public.AddConstraints("POST", "/reset-password", kern.Constraints{Validate: xvalidator.BodyValidator[application.
		ResetPasswordRequest]()}, h.ResetPassword)

	protected := app.Group("/api/v1/auth", jwtAuth)
	protected.GET("/me", h.Me)
	protected.POST("/logout", h.Logout)
	protected.AddConstraints("POST", "/change-password", kern.Constraints{Validate: xvalidator.
		BodyValidator[application.ChangePasswordRequest]()}, h.ChangePassword)
	protected.AddConstraints("PUT", "/profile", kern.Constraints{Validate: xvalidator.BodyValidator[application.
		UpdateProfileRequest]()}, h.UpdateProfile)
}
