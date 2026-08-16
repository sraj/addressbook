package shared

import (
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mobentum/xdb"
)

func NewDB(driver, dsn string) (*xdb.DB, error) {
	return xdb.New(xdb.DBConfig{
		Driver: driver,
		DSN:    dsn,
	})
}
