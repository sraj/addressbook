package interfaces

import "github.com/mobentum/kern"

func (h *Handler) RegisterRoutes(app *kern.App, jwtAuth kern.MiddlewareFunc) {
	g := app.Group("/api/v1/admin", jwtAuth)
	g.GET("/users", h.ListUsers)
	g.PUT("/users/{id}/status", h.UpdateStatus)
	g.GET("/stats", h.Stats)
	g.GET("/plans", h.ListPlans)
	g.PUT("/plans/{id}/price-id", h.UpdatePriceID)
	g.POST("/plans/sync-prices", h.SyncPrices)
}
