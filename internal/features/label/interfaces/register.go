package interfaces

import (
	"github.com/mobentum/xdb"
	"github.com/samber/do/v2"
	billingInfra "github.com/sraj/addressbook/internal/billing/infrastructure"
	"github.com/sraj/addressbook/internal/config"
	contactApp "github.com/sraj/addressbook/internal/features/contact/application"
	"github.com/sraj/addressbook/internal/features/label/application"
	"github.com/sraj/addressbook/internal/features/label/infrastructure"
)

// Provide registers the label service and handler into the injector.
func Provide(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*application.Service, error) {
		db := do.MustInvoke[*xdb.DB](i)
		cfg := do.MustInvoke[*config.Config](i)
		contacts := do.MustInvoke[*contactApp.Service](i)

		var stripe application.CheckoutCreator
		if v, err := do.Invoke[*billingInfra.StripeService](i); err == nil {
			stripe = v
		}

		return application.NewService(
			infrastructure.NewSQLiteRepo(db),
			contacts,
			stripe,
			cfg.AppURL,
			application.Options{
				PriceCents:     cfg.LabelPriceCents,
				Currency:       cfg.LabelCurrency,
				LabelsPerSheet: cfg.LabelLabelsPerSheet,
			},
		), nil
	})

	do.Provide(i, func(i do.Injector) (*Handler, error) {
		svc := do.MustInvoke[*application.Service](i)
		return NewHandler(svc), nil
	})
}
