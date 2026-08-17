package interfaces

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/mobentum/kern"
	"github.com/mobentum/kern/extensions/xvalidator"
	billingDomain "github.com/sraj/addressbook/internal/billing/domain"
	collectionDomain "github.com/sraj/addressbook/internal/features/collection/domain"
	contactDomain "github.com/sraj/addressbook/internal/features/contact/domain"
	"github.com/sraj/addressbook/internal/shared"
)

type Handler struct {
	svc collectionService
}

func NewHandler(svc collectionService) *Handler {
	return &Handler{svc: svc}
}

// ── Protected: CRUD ──────────────────────────────────────────────

func (h *Handler) List(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}
	collections, err := h.svc.List(c.Context(), userID)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to list collections", err)
		return
	}
	_ = c.JSON(http.StatusOK, map[string][]collectionDomain.Collection{"collections": collections})
}

func (h *Handler) Create(c *kern.Context) {
	req, ok := xvalidator.Validated[CreateRequest](c.Context())
	if !ok {
		return
	}
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}
	collection, err := h.svc.Create(c.Context(), userID, req.Name)
	if err != nil {
		if errors.Is(err, collectionDomain.ErrCollectionName) {
			shared.SendError(c, http.StatusBadRequest, err.Error())
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to create collection", err)
		return
	}
	_ = c.JSON(http.StatusCreated, collection)
}

func (h *Handler) Get(c *kern.Context) {
	id, userID, ok := parseIDAndUser(c)
	if !ok {
		return
	}
	collection, err := h.svc.Get(c.Context(), id, userID)
	if err != nil {
		if errors.Is(err, collectionDomain.ErrCollectionNotFound) {
			shared.SendError(c, http.StatusNotFound, "collection not found")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to get collection", err)
		return
	}
	_ = c.JSON(http.StatusOK, collection)
}

func (h *Handler) Rename(c *kern.Context) {
	id, userID, ok := parseIDAndUser(c)
	if !ok {
		return
	}
	req, ok := xvalidator.Validated[CreateRequest](c.Context())
	if !ok {
		return
	}
	collection, err := h.svc.Rename(c.Context(), id, userID, req.Name)
	if err != nil {
		if errors.Is(err, collectionDomain.ErrCollectionNotFound) {
			shared.SendError(c, http.StatusNotFound, "collection not found")
			return
		}
		if errors.Is(err, collectionDomain.ErrCollectionName) {
			shared.SendError(c, http.StatusBadRequest, err.Error())
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to rename collection", err)
		return
	}
	_ = c.JSON(http.StatusOK, collection)
}

func (h *Handler) Delete(c *kern.Context) {
	id, userID, ok := parseIDAndUser(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Context(), id, userID); err != nil {
		if errors.Is(err, collectionDomain.ErrCollectionNotFound) {
			shared.SendError(c, http.StatusNotFound, "collection not found")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to delete collection", err)
		return
	}
	c.NoContent(http.StatusNoContent)
}

func (h *Handler) RegenerateToken(c *kern.Context) {
	id, userID, ok := parseIDAndUser(c)
	if !ok {
		return
	}
	collection, err := h.svc.RegenerateToken(c.Context(), id, userID)
	if err != nil {
		if errors.Is(err, collectionDomain.ErrCollectionNotFound) {
			shared.SendError(c, http.StatusNotFound, "collection not found")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to regenerate token", err)
		return
	}
	_ = c.JSON(http.StatusOK, collection)
}

func (h *Handler) ListContacts(c *kern.Context) {
	id, userID, ok := parseIDAndUser(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	contacts, total, err := h.svc.ListContacts(c.Context(), userID, id, page, size)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to list contacts", err)
		return
	}
	_ = c.JSON(http.StatusOK, shared.NewPaginatedResponse(contacts, total, page, size))
}

// AddContact creates a new contact inside this collection (authenticated).
func (h *Handler) AddContact(c *kern.Context) {
	id, userID, ok := parseIDAndUser(c)
	if !ok {
		return
	}
	req, ok := xvalidator.Validated[ContactRequest](c.Context())
	if !ok {
		return
	}

	contact, err := h.svc.AddContact(c.Context(), userID, id, req.Name, req.Emails, req.Phones, req.ToAddresses(), req.Notes)
	if err != nil {
		if errors.Is(err, collectionDomain.ErrCollectionNotFound) {
			shared.SendError(c, http.StatusNotFound, "collection not found")
			return
		}
		if errors.Is(err, billingDomain.ErrQuotaExceeded) {
			shared.SendError(c, http.StatusForbidden, err.Error())
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to add contact", err)
		return
	}
	_ = c.JSON(http.StatusCreated, contact)
}

// MoveContact assigns an existing contact to this collection.
func (h *Handler) MoveContact(c *kern.Context) {
	id, contactID, userID, ok := parseCollectionAndContact(c)
	if !ok {
		return
	}
	if err := h.svc.MoveContact(c.Context(), userID, id, contactID); err != nil {
		if errors.Is(err, collectionDomain.ErrCollectionNotFound) {
			shared.SendError(c, http.StatusNotFound, "collection not found")
			return
		}
		if errors.Is(err, contactDomain.ErrContactNotFound) {
			shared.SendError(c, http.StatusNotFound, "contact not found")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to move contact", err)
		return
	}
	c.NoContent(http.StatusNoContent)
}

// RemoveContact clears a contact's membership from this collection.
func (h *Handler) RemoveContact(c *kern.Context) {
	id, contactID, userID, ok := parseCollectionAndContact(c)
	if !ok {
		return
	}
	if err := h.svc.RemoveContact(c.Context(), userID, id, contactID); err != nil {
		if errors.Is(err, collectionDomain.ErrCollectionNotFound) {
			shared.SendError(c, http.StatusNotFound, "collection not found")
			return
		}
		if errors.Is(err, contactDomain.ErrContactNotFound) {
			shared.SendError(c, http.StatusNotFound, "contact not found")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to remove contact", err)
		return
	}
	c.NoContent(http.StatusNoContent)
}

// ── Public: invite page ─────────────────────────────────────────

func (h *Handler) PublicInfo(c *kern.Context) {
	token := strings.TrimSpace(c.Param("token"))
	if token == "" {
		shared.SendError(c, http.StatusBadRequest, "missing invite token")
		return
	}
	collection, err := h.svc.GetByToken(c.Context(), token)
	if err != nil {
		if errors.Is(err, collectionDomain.ErrCollectionNotFound) {
			shared.SendError(c, http.StatusNotFound, "invite link is invalid or has been deactivated")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to load invite", err)
		return
	}
	_ = c.JSON(http.StatusOK, map[string]interface{}{
		"name":  collection.Name,
		"token": collection.InviteToken,
	})
}

func (h *Handler) PublicSubmit(c *kern.Context) {
	token := strings.TrimSpace(c.Param("token"))
	if token == "" {
		shared.SendError(c, http.StatusBadRequest, "missing invite token")
		return
	}
	req, ok := xvalidator.Validated[SubmitRequest](c.Context())
	if !ok {
		return
	}

	contact, err := h.svc.SubmitAddress(c.Context(), token, req.Name, req.Email, req.Phone, req.Address.ToDomain())
	if err != nil {
		if errors.Is(err, collectionDomain.ErrInvalidToken) {
			shared.SendError(c, http.StatusNotFound, "invite link is invalid or has been deactivated")
			return
		}
		if errors.Is(err, billingDomain.ErrQuotaExceeded) {
			shared.SendError(c, http.StatusForbidden, "this address book is full")
			return
		}
		shared.SendError(c, http.StatusInternalServerError, "failed to save your address", err)
		return
	}
	_ = c.JSON(http.StatusCreated, map[string]interface{}{
		"status": "ok",
		"name":   contact.Name,
	})
}

func parseIDAndUser(c *kern.Context) (uint, uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid collection id")
		return 0, 0, false
	}
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return 0, 0, false
	}
	return uint(id), userID, true
}

func parseCollectionAndContact(c *kern.Context) (collectionID, contactID, userID uint, ok bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid collection id")
		return 0, 0, 0, false
	}
	cid, err := strconv.ParseUint(c.Param("contactID"), 10, 64)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "invalid contact id")
		return 0, 0, 0, false
	}
	uid, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return 0, 0, 0, false
	}
	return uint(id), uint(cid), uid, true
}
