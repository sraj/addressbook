package interfaces

import (
	"github.com/mobentum/kern"
	"github.com/mobentum/kern/extensions/xvalidator"
	"github.com/sraj/addressbook/internal/features/bookmark/application"
)

func (h *Handler) RegisterRoutes(app *kern.App, jwtAuth kern.MiddlewareFunc) {
	g := app.Group("/api/v1/bookmarks", jwtAuth)
	g.GET("", h.List)
	g.AddConstraints("POST", "", kern.Constraints{Validate: xvalidator.BodyValidator[application.CreateRequest]()}, h.Create)
	g.GET("/{id}", h.Get)
	g.AddConstraints("PUT", "/{id}", kern.Constraints{Validate: xvalidator.BodyValidator[application.UpdateRequest]()}, h.Update)
	g.DELETE("/{id}", h.Delete)
	g.AddConstraints("POST", "/import", kern.Constraints{Validate: xvalidator.BodyValidator[application.ImportRequest]()}, h.Import)
	g.POST("/import-html", h.ImportHTML)
	g.GET("/categories", h.ListCategories)
}
