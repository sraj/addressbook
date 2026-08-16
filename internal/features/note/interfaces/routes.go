package interfaces

import (
	"github.com/mobentum/kern"
	"github.com/mobentum/kern/extensions/xvalidator"
	"github.com/sraj/addressbook/internal/features/note/application"
)

func (h *Handler) RegisterRoutes(app *kern.App, jwtAuth kern.MiddlewareFunc) {
	g := app.Group("/api/v1/notes", jwtAuth)
	g.GET("", h.List)
	g.AddConstraints("POST", "", kern.Constraints{Validate: xvalidator.BodyValidator[application.CreateRequest]()}, h.Create)
	g.GET("/{id}", h.Get)
	g.AddConstraints("PUT", "/{id}", kern.Constraints{Validate: xvalidator.BodyValidator[application.UpdateRequest]()}, h.Update)
	g.DELETE("/{id}", h.Delete)
}
