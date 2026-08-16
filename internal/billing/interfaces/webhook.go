package interfaces

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/mobentum/kern"
	"github.com/sraj/addressbook/internal/mailer"
	"github.com/sraj/addressbook/internal/shared"
	"github.com/stripe/stripe-go/v79"
)

func (h *Handler) Webhook(c *kern.Context) {
	if h.stripe == nil {
		c.NoContent(http.StatusNotImplemented)
		return
	}

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "failed to read request body")
		return
	}

	sigHeader := c.Request.Header.Get("Stripe-Signature")
	event, err := h.stripe.VerifyWebhook(payload, sigHeader)
	if err != nil {
		shared.SendError(c, http.StatusBadRequest, "webhook verification failed", err)
		return
	}

	slog.Info("webhook event received", "type", event.Type, "id", event.ID)
	go h.processWebhookEvent(event)
	c.NoContent(http.StatusOK)
}

func (h *Handler) processWebhookEvent(event stripe.Event) {
	ctx := context.Background()

	switch event.Type {
	case "checkout.session.completed", "checkout_session.completed":
		h.handleCheckoutCompleted(ctx, event)
	case "customer.subscription.updated", "customer_subscription.updated":
		h.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted", "customer_subscription.deleted":
		h.handleSubscriptionDeleted(ctx, event)
	case "invoice.payment_succeeded", "invoice.paid", "invoice_payment.paid":
		h.handleInvoicePaid(ctx, event)
	case "invoice.payment_failed", "invoice_payment.failed":
		h.handleInvoiceFailed(ctx, event)
	default:
		slog.Info("unhandled webhook event", "type", event.Type)
	}
}

func (h *Handler) handleCheckoutCompleted(ctx context.Context, event stripe.Event) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		slog.Error("failed to unmarshal checkout session", "error", err)
		return
	}

	accountIDStr := session.Metadata["account_id"]

	var accountID uint
	if _, err := fmt.Sscanf(accountIDStr, "%d", &accountID); err != nil {
		slog.Error("invalid account_id in metadata", "account_id", accountIDStr)
		return
	}

	account, err := h.svc.GetAccountByID(ctx, accountID)
	if err != nil {
		slog.Error("account not found for checkout", "account_id", accountID)
		return
	}

	if session.Customer != nil && account.StripeCustomerID == "" {
		h.svc.SetStripeCustomerID(ctx, account.ID, session.Customer.ID)
	}

	// Determine which plan was purchased
	planName := session.Metadata["plan_name"]
	if planName == "" {
		planName = "pro"
	}
	plan, err := h.svc.GetPlanByName(ctx, planName)
	if err != nil {
		slog.Error("plan not found", "plan_name", planName)
		return
	}

	stripeSubID := ""
	if session.Subscription != nil {
		stripeSubID = session.Subscription.ID
	}

	if err := h.svc.SetPlanWithStripeID(ctx, account.ID, plan.ID, stripeSubID); err != nil {
		slog.Error("failed to set plan", "account_id", account.ID, "plan", plan.Name, "error", err)
		return
	}
	slog.Info("account upgraded", "account_id", account.ID, "plan", plan.Name)

	h.trySendEmail(account.Name, "Upgrade confirmed!",
		mailer.SubscriptionConfirmedEmail(plan.Name, h.appURL+"/settings/billing"))
}

func (h *Handler) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		slog.Error("failed to unmarshal subscription", "error", err)
		return
	}

	customerID := sub.Customer.ID
	if customerID == "" {
		return
	}

	// Sync plan from subscription items
	var priceID string
	if len(sub.Items.Data) > 0 {
		priceID = sub.Items.Data[0].Price.ID
	}

	switch sub.Status {
	case "past_due":
		h.trySendEmail("", "Payment failed",
			mailer.PaymentFailedEmail(h.appURL+"/settings/billing"))

	case "active":
		// Plan may have changed via Stripe portal — sync it
		if priceID != "" {
			account, err := h.svc.GetAccountByCustomerID(ctx, customerID)
			if err != nil {
				slog.Error("account not found for subscription update", "customer", customerID)
				return
			}

			// Find matching plan by stripe_price_id
			plans, err := h.svc.GetPlans(ctx)
			if err != nil {
				slog.Error("failed to list plans", "error", err)
				return
			}

			for _, p := range plans {
				if p.StripePriceID == priceID {
					if p.ID != account.PlanID {
						h.svc.SetPlan(ctx, account.ID, p.ID)
						slog.Info("plan synced from Stripe", "account_id", account.ID, "plan", p.Name)
					}
					break
				}
			}
		}

		h.trySendEmail("", "Payment received",
			mailer.SubscriptionConfirmedEmail("Pro", h.appURL+"/settings/billing"))
	}
}

func (h *Handler) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		slog.Error("failed to unmarshal subscription", "error", err)
		return
	}

	customerID := sub.Customer.ID
	if customerID == "" {
		slog.Error("subscription has no customer ID")
		return
	}

	account, err := h.svc.GetAccountByCustomerID(ctx, customerID)
	if err != nil {
		slog.Error("account not found for subscription", "customer", customerID)
		return
	}

	freePlan, err := h.svc.GetPlanByName(ctx, "free")
	if err != nil {
		slog.Error("free plan not found", "error", err)
		return
	}

	if err := h.svc.SetPlan(ctx, account.ID, freePlan.ID); err != nil {
		slog.Error("failed to downgrade account", "account_id", account.ID, "error", err)
		return
	}

	slog.Info("account downgraded to free", "account_id", account.ID)
	h.trySendEmail(account.Name, "Plan downgraded",
		mailer.SubscriptionCanceledEmail("Pro", h.appURL+"/settings/billing"))
}

func (h *Handler) handleInvoicePaid(ctx context.Context, event stripe.Event) {
	var inv stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		slog.Error("failed to unmarshal invoice", "error", err)
		return
	}

	if inv.HostedInvoiceURL != "" {
		amount := fmt.Sprintf("$%.2f", float64(inv.AmountPaid)/100)
		h.trySendEmail("", "Invoice paid",
			mailer.InvoiceEmail(amount, inv.HostedInvoiceURL))
	}
}

func (h *Handler) handleInvoiceFailed(ctx context.Context, event stripe.Event) {
	var inv stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		slog.Error("failed to unmarshal invoice", "error", err)
		return
	}

	h.trySendEmail("", "Payment failed",
		mailer.PaymentFailedEmail(h.appURL+"/settings/billing"))
}

func (h *Handler) trySendEmail(to, subject, body string) {
	if h.mailer == nil {
		return
	}
	go func() {
		if err := h.mailer.Send(to, subject, body); err != nil {
			slog.Error("failed to send email", "error", err, "to", to, "subject", subject)
		}
	}()
}
