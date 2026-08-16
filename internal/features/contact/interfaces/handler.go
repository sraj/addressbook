package interfaces

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/mobentum/kern"
	"github.com/mobentum/kern/extensions/xvalidator"
	billingDomain "github.com/sraj/addressbook/internal/billing/domain"
	"github.com/sraj/addressbook/internal/features/contact/application"
	"github.com/sraj/addressbook/internal/features/contact/domain"
	"github.com/sraj/addressbook/internal/shared"
)

type Handler struct {
	svc contactService
}

func NewHandler(svc contactService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	q := c.DefaultQuery("q", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	contacts, total, err := h.svc.List(c.Context(), userID, q, page, size)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to list contacts", err)
		return
	}

	_ = c.JSON(http.StatusOK, shared.NewPaginatedResponse(contacts, total, page, size))
}

func (h *Handler) Get(c *kern.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid contact id")
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	contact, err := h.svc.GetByID(c.Context(), uint(id), userID)
	if err != nil {
		if err == domain.ErrContactNotFound {
			shared.SendError(c, http.StatusNotFound, "contact not found")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to get contact", err)
		return
	}

	_ = c.JSON(http.StatusOK, contact)
}

func (h *Handler) Create(c *kern.Context) {
	req, ok := xvalidator.Validated[application.CreateRequest](c.Context())
	if !ok {
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	addresses := make([]domain.Address, len(req.Addresses))
	for i, a := range req.Addresses {
		addresses[i] = a.ToDomain()
	}

	contact, err := h.svc.Create(c.Context(), userID, req.Name, req.Emails, req.Phones, addresses, req.Notes)
	if err != nil {
		if errors.Is(err, billingDomain.ErrQuotaExceeded) {
			shared.SendError(c, http.StatusForbidden, err.Error())
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to create contact", err)
		return
	}

	_ = c.JSON(http.StatusCreated, contact)
}

func (h *Handler) Update(c *kern.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid contact id")
		return
	}

	req, ok := xvalidator.Validated[application.UpdateRequest](c.Context())
	if !ok {
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	addresses := make([]domain.Address, len(req.Addresses))
	for i, a := range req.Addresses {
		addresses[i] = a.ToDomain()
	}

	contact, err := h.svc.Update(c.Context(), uint(id), userID, req.Name, req.Emails, req.Phones, addresses, req.Notes)
	if err != nil {
		if err == domain.ErrContactNotFound {
			shared.SendError(c, http.StatusNotFound, "contact not found")
			return
		}
		if l := kern.LoggerFromContext(c.Context()); l != nil {
			l.Error("failed to update contact", "error", err, "contact_id", id)
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to update contact", err)
		return
	}

	_ = c.JSON(http.StatusOK, contact)
}

func (h *Handler) Delete(c *kern.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid contact id")
		return
	}

	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}

	if err := h.svc.Delete(c.Context(), uint(id), userID); err != nil {
		if err == domain.ErrContactNotFound {
			shared.SendError(c, http.StatusNotFound, "contact not found")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to delete contact", err)
		return
	}

	c.NoContent(http.StatusNoContent)
}
