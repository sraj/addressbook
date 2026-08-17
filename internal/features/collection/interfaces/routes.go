package interfaces

import (
	"github.com/mobentum/kern"
	"github.com/mobentum/kern/extensions/xvalidator"
)

func (h *Handler) RegisterRoutes(app *kern.App, jwtAuth kern.MiddlewareFunc) {
	// Authenticated collection management.
	g := app.Group("/api/v1/collections", jwtAuth)
	g.GET("", h.List)
	g.AddConstraints("POST", "", kern.Constraints{Validate: xvalidator.BodyValidator[CreateRequest]()}, h.Create)
	g.GET("/{id}", h.Get)
	g.AddConstraints("PUT", "/{id}", kern.Constraints{Validate: xvalidator.BodyValidator[CreateRequest]()}, h.Rename)
	g.DELETE("/{id}", h.Delete)
	g.POST("/{id}/regenerate-token", h.RegenerateToken)
	g.GET("/{id}/contacts", h.ListContacts)
	g.AddConstraints("POST", "/{id}/contacts", kern.Constraints{Validate: xvalidator.BodyValidator[ContactRequest]()}, h.AddContact)
	g.PUT("/{id}/contacts/{contactID}", h.MoveContact)
	g.DELETE("/{id}/contacts/{contactID}", h.RemoveContact)

	// Public invite endpoints — no auth, reachable by anyone with the token.
	pub := app.Group("/api/v1/invites")
	pub.GET("/{token}", h.PublicInfo)
	pub.AddConstraints("POST", "/{token}", kern.Constraints{Validate: xvalidator.BodyValidator[SubmitRequest]()}, h.PublicSubmit)
}
