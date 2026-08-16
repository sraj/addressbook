package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	authInfra "github.com/sraj/addressbook/internal/auth/infrastructure"
	billingInfra "github.com/sraj/addressbook/internal/billing/infrastructure"
	"github.com/sraj/addressbook/internal/config"
	"github.com/sraj/addressbook/internal/shared"
	"github.com/sraj/addressbook/migrations"
	"github.com/mobentum/kern/extensions/xlog"
	"github.com/mobentum/xdb"
	"github.com/urfave/cli/v2"
)

func main() {
	slogger := xlog.NewLogger(xlog.Config{Format: "json", Level: slog.LevelInfo})
	slog.SetDefault(slogger)

	cfg, err := config.Load()
	if err != nil {
		slogger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	app := &cli.App{
		Name:  "addressbook",
		Usage: "Admin CLI for addressbook",
		Commands: []*cli.Command{
			{
				Name:  "migrate:up",
				Usage: "Run database migrations",
				Action: func(c *cli.Context) error {
					db := openDB(cfg.DatabaseDriver, cfg.DatabasePath)
					defer func() { _ = db.Close() }()
					if err := migrations.MigrateUp(db, cfg.DatabaseDriver); err != nil {
						slogger.Error("migration failed", "error", err, "msg", err.Error())
						return err
					}
					slogger.Info("migrations complete")
					return nil
				},
			},
			{
				Name:  "migrate:down",
				Usage: "Roll back all migrations",
				Action: func(c *cli.Context) error {
					db := openDB(cfg.DatabaseDriver, cfg.DatabasePath)
					defer func() { _ = db.Close() }()
					if err := migrations.MigrateDown(db, cfg.DatabaseDriver); err != nil {
						slogger.Error("rollback failed", "error", err)
						return err
					}
					slogger.Info("rollback complete")
					return nil
				},
			},
			{
				Name:  "seed",
				Usage: "Seed the database with initial data",
				Action: func(c *cli.Context) error {
					ctx := context.Background()
					db := openDB(cfg.DatabaseDriver, cfg.DatabasePath)
					defer func() { _ = db.Close() }()

					if err := authInfra.Seed(ctx, db); err != nil {
						slogger.Error("auth seed failed", "error", err)
						return err
					}
					slogger.Info("auth seed complete")

					if err := billingInfra.Seed(ctx, db, cfg.StripeSecretKey); err != nil {
						slogger.Error("billing seed failed", "error", err)
						return err
					}
					slogger.Info("billing seed complete")

					return nil
				},
			},
			{
				Name:  "setup",
				Usage: "Run migrations then seed",
				Action: func(c *cli.Context) error {
					ctx := context.Background()
					db := openDB(cfg.DatabaseDriver, cfg.DatabasePath)
					defer func() { _ = db.Close() }()

					if err := migrations.MigrateUp(db, cfg.DatabaseDriver); err != nil {
						slogger.Error("migration failed", "error", err, "msg", err.Error())
						return err
					}
					slogger.Info("migrations complete")

					if err := authInfra.Seed(ctx, db); err != nil {
						slogger.Error("auth seed failed", "error", err)
						return err
					}
					slogger.Info("auth seed complete")

					if err := billingInfra.Seed(ctx, db, cfg.StripeSecretKey); err != nil {
						slogger.Error("billing seed failed", "error", err)
						return err
					}
					slogger.Info("billing seed complete")

					return nil
				},
			},
			{
				Name:  "db:clean",
				Usage: "Drop and re-create the database, then run migrations",
				Action: func(c *cli.Context) error {
					if err := dropRecreateDB(cfg.DatabaseDriver, cfg.DatabasePath); err != nil {
						slogger.Error("db clean failed", "error", err)
						return err
					}
					db := openDB(cfg.DatabaseDriver, cfg.DatabasePath)
					defer func() { _ = db.Close() }()
					if err := migrations.MigrateUp(db, cfg.DatabaseDriver); err != nil {
						slogger.Error("migration failed", "error", err, "msg", err.Error())
						return err
					}
					slogger.Info("database cleaned and migrated")
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		slogger.Error("command failed", "error", err)
		os.Exit(1)
	}
}

func openDB(driver, dsn string) *xdb.DB {
	db, err := shared.NewDB(driver, dsn)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	return db
}

func dropRecreateDB(driver, dsn string) error {
	switch driver {
	case "sqlite3":
		if err := os.Remove(dsn); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove sqlite db: %w", err)
		}
		return nil
	case "postgres":
		u, err := url.Parse(dsn)
		if err != nil {
			return fmt.Errorf("parse postgres dsn: %w", err)
		}
		dbName := strings.TrimPrefix(u.Path, "/")
		if dbName == "" {
			return errors.New("postgres dsn missing database name")
		}
		u.Path = "/postgres"
		conn, err := sql.Open("postgres", u.String())
		if err != nil {
			return err
		}
		defer func() { _ = conn.Close() }()

		if _, err := conn.Exec("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()", dbName); err != nil {
			return fmt.Errorf("terminate connections: %w", err)
		}
		if _, err := conn.Exec("DROP DATABASE IF EXISTS " + quoteIdent(dbName)); err != nil {
			return fmt.Errorf("drop database: %w", err)
		}
		if _, err := conn.Exec("CREATE DATABASE " + quoteIdent(dbName)); err != nil {
			return fmt.Errorf("create database: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported database driver %q", driver)
	}
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
