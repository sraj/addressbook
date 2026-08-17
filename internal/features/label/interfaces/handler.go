package interfaces

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/mobentum/kern"
	"github.com/mobentum/kern/extensions/xvalidator"
	labelDomain "github.com/sraj/addressbook/internal/features/label/domain"
	"github.com/sraj/addressbook/internal/shared"
)

type Handler struct {
	svc labelService
}

func NewHandler(svc labelService) *Handler {
	return &Handler{svc: svc}
}

// Sheet renders a printable HTML label sheet for the requested Avery template.
func (h *Handler) Sheet(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}
	var collectionID uint
	if v, err := strconv.ParseUint(c.DefaultQuery("collection_id", "0"), 10, 64); err == nil {
		collectionID = uint(v)
	}
	format := c.DefaultQuery("format", labelDomain.DefaultLabelFormat)

	html, err := h.svc.GenerateSheet(c.Context(), userID, collectionID, format)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "failed to generate label sheet", err)
		return
	}
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	c.SetHeader("X-Robots-Tag", "noindex")
	// The sheet embeds an inline print script (see labelDomain.PrintScript).
	// Allow it via its sha256 hash rather than unsafe-inline, and keep the
	// remaining directives aligned with the app-wide policy.
	sum := sha256.Sum256([]byte(labelDomain.PrintScript))
	scriptHash := "sha256-" + base64.StdEncoding.EncodeToString(sum[:])
	c.SetHeader("Content-Security-Policy",
		"default-src 'self'; script-src 'self' '"+scriptHash+"'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
	_ = c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// Order creates a Stripe checkout for printed labels.
func (h *Handler) Order(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}
	req, ok := xvalidator.Validated[OrderRequest](c.Context())
	if !ok {
		return
	}

	order, url, err := h.svc.CreateOrder(c.Context(), userID, req.CollectionID, req.Email, req.Format)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, err.Error())
		return
	}
	_ = c.JSON(http.StatusCreated, map[string]interface{}{
		"order": order,
		"url":   url,
	})
}

// Formats lists the supported label templates.
func (h *Handler) Formats(c *kern.Context) {
	_ = c.JSON(http.StatusOK, map[string]interface{}{"formats": h.svc.SupportedFormats()})
}

func (h *Handler) ListOrders(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}
	orders, err := h.svc.ListOrders(c.Context(), userID)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to list orders", err)
		return
	}
	_ = c.JSON(http.StatusOK, map[string]interface{}{"orders": orders})
}

func (h *Handler) Confirm(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}
	sessionID := c.DefaultQuery("session_id", "")
	if sessionID == "" {
		shared.SendError(c, http.StatusBadRequest, "missing session_id")
		return
	}
	order, err := h.svc.ConfirmOrder(c.Context(), userID, sessionID)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "failed to confirm order", err)
		return
	}
	_ = c.JSON(http.StatusOK, order)
}
