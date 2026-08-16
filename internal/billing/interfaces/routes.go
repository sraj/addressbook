package interfaces

import "github.com/mobentum/kern"

func (h *Handler) RegisterRoutes(app *kern.App, jwtAuth kern.MiddlewareFunc) {
	g := app.Group("/api/v1/billing", jwtAuth)
	g.GET("/plans", h.ListPlans)
	g.GET("/usage", h.Usage)
	g.GET("/invoices", h.Invoices)
	g.POST("/checkout", h.Checkout)
	g.POST("/portal", h.Portal)
	g.POST("/change-plan", h.ChangePlan)
	g.POST("/cancel", h.Cancel)

	app.POST("/api/v1/webhook/stripe", h.Webhook)
}
