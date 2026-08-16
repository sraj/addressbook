package interfaces

import (
	"net/http"

	"github.com/mobentum/kern"
	"github.com/sraj/addressbook/internal/billing/domain"
	"github.com/sraj/addressbook/internal/billing/infrastructure"
	"github.com/sraj/addressbook/internal/shared"
)

type Handler struct {
	svc    billingService
	stripe stripeChecker
	mailer mailSender
	appURL string
}

func NewHandler(svc billingService, stripe stripeChecker, mailer mailSender, appURL string) *Handler {
	return &Handler{svc: svc, stripe: stripe, mailer: mailer, appURL: appURL}
}

type UsageResponse struct {
	Usage  map[string]int `json:"usage"`
	Plan   *domain.Plan   `json:"plan"`
	Limits map[string]int `json:"limits"`
}

func (h *Handler) ListPlans(c *kern.Context) {
	plans, err := h.svc.GetPlans(c.Context())
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to list plans", err)
		return
	}
	_ = c.JSON(http.StatusOK, map[string][]domain.Plan{"plans": plans})
}

func (h *Handler) Usage(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}
	usage, plan, err := h.svc.GetUsage(c.Context(), userID)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to get usage", err)
		return
	}
	_ = c.JSON(http.StatusOK, UsageResponse{
		Usage:  usage,
		Plan:   plan,
		Limits: plan.Limits(),
	})
}

func (h *Handler) Checkout(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}
	if h.stripe == nil {
		shared.SendError(c, http.StatusBadRequest, "billing is not configured")
		return
	}

	planName := c.DefaultQuery("plan", "pro")
	account, err := h.svc.GetAccount(c.Context(), userID)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "account not found", err)
		return
	}

	plan, err := h.svc.GetPlanByName(c.Context(), planName)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "plan not found")
		return
	}
	if plan.StripePriceID == "" {
		shared.SendError(c, http.StatusBadRequest, "plan has no Stripe price ID configured")
		return
	}

	url, err := h.stripe.CreateCheckoutSession(c.Context(), account, account.Name, plan.StripePriceID, plan.Name)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to create checkout session", err)
		return
	}
	if account.StripeCustomerID != "" {
		if err := h.svc.SetStripeCustomerID(c.Context(), account.ID, account.StripeCustomerID); err != nil {
			shared.SendError(c, http.StatusInternalServerError, "failed to save customer", err)
			return
		}
	}
	_ = c.JSON(http.StatusOK, map[string]string{"url": url})
}

func (h *Handler) Portal(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}
	if h.stripe == nil {
		shared.SendError(c, http.StatusBadRequest, "billing is not configured")
		return
	}
	account, err := h.svc.GetAccount(c.Context(), userID)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "account not found", err)
		return
	}
	url, err := h.stripe.CreatePortalSession(c.Context(), account)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to create portal session", err)
		return
	}
	if account.StripeCustomerID != "" {
		if err := h.svc.SetStripeCustomerID(c.Context(), account.ID, account.StripeCustomerID); err != nil {
			shared.SendError(c, http.StatusInternalServerError, "failed to save customer", err)
			return
		}
	}
	_ = c.JSON(http.StatusOK, map[string]string{"url": url})
}

func (h *Handler) Cancel(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}
	if h.stripe == nil {
		shared.SendError(c, http.StatusBadRequest, "billing is not configured")
		return
	}

	account, err := h.svc.GetAccount(c.Context(), userID)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "account not found", err)
		return
	}

	sub, err := h.svc.GetSubscription(c.Context(), account.ID)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to get subscription", err)
		return
	}
	if sub == nil || sub.StripeSubscriptionID == "" {
		shared.SendError(c, http.StatusBadRequest, "no active subscription to cancel")
		return
	}

	if err := h.stripe.CancelSubscription(c.Context(), sub.StripeSubscriptionID); err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to cancel subscription", err)
		return
	}

	_ = c.JSON(http.StatusOK, map[string]string{"status": "canceled"})
}

func (h *Handler) ChangePlan(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}
	if h.stripe == nil {
		shared.SendError(c, http.StatusBadRequest, "billing is not configured")
		return
	}

	planName := c.DefaultQuery("plan", "")
	if planName == "" {
		shared.SendError(c, http.StatusBadRequest, "plan is required")
		return
	}

	plan, err := h.svc.GetPlanByName(c.Context(), planName)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "plan not found")
		return
	}
	if plan.StripePriceID == "" {
		shared.SendError(c, http.StatusBadRequest, "plan has no Stripe price ID configured")
		return
	}

	// Get the user's existing Stripe subscription
	account, err := h.svc.GetAccount(c.Context(), userID)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "account not found", err)
		return
	}
	sub, err := h.svc.GetSubscription(c.Context(), account.ID)
	if err != nil || sub == nil || sub.StripeSubscriptionID == "" {
		shared.SendError(c, http.StatusBadRequest, "no active subscription found")
		return
	}

	if err := h.stripe.ChangeSubscriptionPlan(c.Context(), sub.StripeSubscriptionID, plan.StripePriceID); err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to change plan", err)
		return
	}

	// Also update our local subscription record
	if err := h.svc.SetPlan(c.Context(), account.ID, plan.ID); err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to update local plan", err)
		return
	}

	_ = c.JSON(http.StatusOK, map[string]string{"status": "changed", "plan": planName, "price_id": plan.StripePriceID})
}

func (h *Handler) Invoices(c *kern.Context) {
	userID, ok := shared.UserIDFromCtx(c.Context())
	if !ok {
		c.NoContent(http.StatusUnauthorized)
		return
	}
	if h.stripe == nil {
		shared.SendError(c, http.StatusBadRequest, "billing is not configured")
		return
	}

	account, err := h.svc.GetAccount(c.Context(), userID)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "account not found", err)
		return
	}

	if account.StripeCustomerID == "" {
		_ = c.JSON(http.StatusOK, []infrastructure.InvoiceRow{})
		return
	}

	invoices, err := h.stripe.ListInvoices(c.Context(), account.StripeCustomerID)
	if err != nil {
		shared.SendError(c, http.StatusInternalServerError, "failed to list invoices", err)
		return
	}

	_ = c.JSON(http.StatusOK, invoices)
}
