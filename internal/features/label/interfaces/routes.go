package interfaces

import (
	"github.com/mobentum/kern"
	"github.com/mobentum/kern/extensions/xvalidator"
)

func (h *Handler) RegisterRoutes(app *kern.App, jwtAuth kern.MiddlewareFunc) {
	g := app.Group("/api/v1/labels", jwtAuth)
	g.GET("/sheet", h.Sheet)
	g.AddConstraints("POST", "/order", kern.Constraints{Validate: xvalidator.BodyValidator[OrderRequest]()}, h.Order)
	g.POST("/confirm", h.Confirm)
	g.GET("/orders", h.ListOrders)
	g.GET("/formats", h.Formats)
}
