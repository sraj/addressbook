package interfaces

import (
	"github.com/sraj/addressbook/internal/billing/application"
	"github.com/sraj/addressbook/internal/billing/infrastructure"
	"github.com/sraj/addressbook/internal/config"
	"github.com/sraj/addressbook/internal/mailer"
	"github.com/mobentum/xdb"
	"github.com/samber/do/v2"
)

// Provide registers the billing service and handler into the injector.
func Provide(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*application.Service, error) {
		db := do.MustInvoke[*xdb.DB](i)
		cfg := do.MustInvoke[*config.Config](i)
		return application.NewService(infrastructure.NewSQLiteRepo(db), cfg.StripeSecretKey), nil
	})

	do.Provide(i, func(i do.Injector) (*Handler, error) {
		svc := do.MustInvoke[*application.Service](i)
		cfg := do.MustInvoke[*config.Config](i)

		var stripe stripeChecker
		if v, err := do.Invoke[*infrastructure.StripeService](i); err == nil {
			stripe = v
		}

		var m mailSender
		if v, err := do.Invoke[*mailer.Mailer](i); err == nil {
			m = v
		}

		return NewHandler(svc, stripe, m, cfg.AppURL), nil
	})
}
