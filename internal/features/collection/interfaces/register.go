package interfaces

import (
	"github.com/mobentum/xdb"
	"github.com/samber/do/v2"
	"github.com/sraj/addressbook/internal/features/collection/application"
	"github.com/sraj/addressbook/internal/features/collection/infrastructure"
	contactApp "github.com/sraj/addressbook/internal/features/contact/application"
)

// Provide registers the collection service and handler into the injector.
func Provide(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*application.Service, error) {
		db := do.MustInvoke[*xdb.DB](i)
		contact := do.MustInvoke[*contactApp.Service](i)
		return application.NewService(infrastructure.NewSQLiteRepo(db), contact), nil
	})

	do.Provide(i, func(i do.Injector) (*Handler, error) {
		svc := do.MustInvoke[*application.Service](i)
		return NewHandler(svc), nil
	})
}
