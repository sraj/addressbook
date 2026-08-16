package interfaces

import (
	"encoding/json"
	"net/http"
	"strconv"

	adminApp "github.com/sraj/addressbook/internal/admin/application"
	"github.com/mobentum/kern"
	"github.com/sraj/addressbook/internal/shared"
)

type Handler struct {
	svc    adminChecker
	plans  planManager
}

func NewHandler(svc adminChecker, plans planManager) *Handler {
	return &Handler{svc: svc, plans: plans}
}

func (h *Handler) requireAdmin(c *kern.Context) bool {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return false
	}
	admin, err := h.svc.IsAdmin(c.Context(), userID)
	if err != nil || !admin {
		shared.SendError(c, http.StatusForbidden, "admin access required")
		return false
	}
	return true
}

func (h *Handler) ListUsers(c *kern.Context) {
	if !h.requireAdmin(c) {
		return
	}
	users, err := h.svc.ListUsers(c.Context())
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to list users", err)
		return
	}
	_ = c.JSON(http.StatusOK, map[string][]adminApp.UserRow{"users": users})
}

func (h *Handler) UpdateStatus(c *kern.Context) {
	if !h.requireAdmin(c) {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid user id")
		return
	}
	status := c.DefaultQuery("status", "active")
	if status != "active" && status != "suspended" {
		shared.SendError(c, http.StatusBadRequest, "status must be active or suspended")
		return
	}
	if err := h.svc.UpdateUserStatus(c.Context(), uint(id), status); err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to update user", err)
		return
	}
	_ = c.JSON(http.StatusOK, map[string]string{"status": status})
}

func (h *Handler) ListPlans(c *kern.Context) {
	if !h.requireAdmin(c) {
		return
	}
	plans, err := h.plans.ListPlans(c.Context())
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to list plans", err)
		return
	}
	_ = c.JSON(http.StatusOK, map[string]any{"plans": plans})
}

func (h *Handler) UpdatePriceID(c *kern.Context) {
	if !h.requireAdmin(c) {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid plan id")
		return
	}
	var req struct {
		PriceID string `json:"stripe_price_id"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.plans.UpdatePlanPriceID(c.Context(), uint(id), req.PriceID); err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to update price ID", err)
		return
	}
	_ = c.JSON(http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) SyncPrices(c *kern.Context) {
	if !h.requireAdmin(c) {
		return
	}
	if err := h.plans.SyncPricesFromStripe(c.Context()); err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to sync prices", err)
		return
	}
	_ = c.JSON(http.StatusOK, map[string]string{"status": "synced"})
}

func (h *Handler) Stats(c *kern.Context) {
	if !h.requireAdmin(c) {
		return
	}
	stats, err := h.svc.Stats(c.Context())
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to get stats", err)
		return
	}
	_ = c.JSON(http.StatusOK, stats)
}
