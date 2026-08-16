package mailer

import (
	"fmt"
)

const baseTemplate = `<!DOCTYPE html>
<html><body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
<div style="background: #f5f5f5; border-radius: 8px; padding: 32px; text-align: center;">
<div style="font-size: 32px; margin-bottom: 8px;">✉️</div>
<h2 style="margin: 0 0 4px;">%s</h2>
<p style="color: #666; margin: 0 0 24px;">%s</p>
<a href="%s" style="display: inline-block; background: #171717; color: #fff; text-decoration: none; padding: 12px 24px; border-radius: 6px; font-size: 14px;">%s</a>
<p style="color: #999; font-size: 12px; margin-top: 24px;">If the button doesn't work, copy this URL into your browser:<br><span style="color: #666;">%s</span></p>
<hr style="border: none; border-top: 1px solid #e5e5e5; margin: 24px 0;">
<p style="color: #999; font-size: 12px;">Address Book — Manage your contacts, notes, and bookmarks</p>
</div></body></html>`

func ResetPasswordEmail(resetLink string) string {
	return fmt.Sprintf(baseTemplate,
		"Reset your password",
		"Click the button below to set a new password.",
		resetLink,
		"Reset Password",
		resetLink,
	)
}

func VerifyEmailEmail(verifyLink string) string {
	return fmt.Sprintf(baseTemplate,
		"Verify your email",
		"Click the button below to verify your email address.",
		verifyLink,
		"Verify Email",
		verifyLink,
	)
}

func WelcomeEmail(name string) string {
	body := `<!DOCTYPE html>
<html><body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
<div style="background: #f5f5f5; border-radius: 8px; padding: 32px; text-align: center;">
<h2 style="margin: 0 0 4px;">Welcome to Address Book!</h2>
<p style="color: #666; margin: 0 0 24px;">Hi ` + name + `, your account is ready. Start managing your contacts, notes, and bookmarks.</p>
<hr style="border: none; border-top: 1px solid #e5e5e5; margin: 24px 0;">
<p style="color: #999; font-size: 12px;">Address Book — Manage your contacts, notes, and bookmarks</p>
</div></body></html>`
	return body
}

func SubscriptionConfirmedEmail(planName string, dashboardLink string) string {
	title := "Upgrade confirmed!"
	desc := fmt.Sprintf("You're now on the %s plan. Enjoy the additional features.", planName)
	return fmt.Sprintf(baseTemplate, title, desc, dashboardLink, "Go to Dashboard", dashboardLink)
}

func SubscriptionCanceledEmail(planName string, dashboardLink string) string {
	title := "Subscription canceled"
	desc := fmt.Sprintf("Your %s plan has been canceled. You'll be downgraded to Free at the end of the billing period.", planName)
	return fmt.Sprintf(baseTemplate, title, desc, dashboardLink, "Go to Dashboard", dashboardLink)
}

func PaymentFailedEmail(retryLink string) string {
	return fmt.Sprintf(baseTemplate,
		"Payment failed",
		"We couldn't process your payment. Please update your billing information to avoid service interruption.",
		retryLink,
		"Update Payment",
		retryLink,
	)
}

func InvoiceEmail(amount string, invoiceLink string) string {
	return fmt.Sprintf(baseTemplate,
		fmt.Sprintf("Invoice — %s", amount),
		"Your receipt is ready. Click below to view or download it.",
		invoiceLink,
		"View Invoice",
		invoiceLink,
	)
}
