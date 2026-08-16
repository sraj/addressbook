package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/sraj/addressbook/internal/app"
	"github.com/sraj/addressbook/internal/config"
	"github.com/sraj/addressbook/internal/shared"
	"github.com/sraj/addressbook/migrations"
	"github.com/mobentum/kern"
	"github.com/mobentum/kern/extensions/xlog"
	"github.com/mobentum/kern/extensions/xotel"
	"github.com/mobentum/kern/middleware"
)

func main() {

	slogger := xlog.NewLogger(xlog.Config{Format: "json", Level: slog.LevelInfo})
	slog.SetDefault(slogger)

	cfg, err := config.Load()
	if err != nil {
		slogger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	shutdownOTel, err := xotel.Setup(xotel.SetupConfig{
		ServiceName: "addressbook",
		Endpoint:    cfg.OTLPEndpoint,
		Insecure:    true,
		Metrics:     true,
	})
	if err != nil {
		slogger.Error("failed to setup OpenTelemetry", "error", err)
		os.Exit(1)
	}
	defer shutdownOTel()

	db, err := shared.NewDB(cfg.DatabaseDriver, cfg.DatabasePath)
	if err != nil {
		slogger.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	if err := migrations.MigrateUp(db, cfg.DatabaseDriver); err != nil {
		slogger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	a := app.New(db, cfg)

	server := kern.New(kern.WithSlogLogger(slogger))

	server.Use(middleware.RequestID())
	server.Use(xotel.Middleware(xotel.Config{ServiceName: "addressbook"}))
	server.Use(kern.Logger(kern.LoggerConfig{
		SLogger: slogger,
		Format:  "json",
		Fields:  map[string]interface{}{"log_type": "access"},
	}))
	server.Use(shared.Recovery())
	server.Use(kern.CORSWithConfig(kern.CORSConfig{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-CSRF-Token"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
	}))
	server.Use(middleware.SecurityHeaders(middleware.SecurityHeadersConfig{
		StrictTransportSecurity: "max-age=31536000; includeSubDomains",
		ContentSecurityPolicy:   "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
		ReferrerPolicy:          "strict-origin-when-cross-origin",
		PermissionsPolicy:       "camera=(), microphone=(), geolocation=(), payment=()",
	}))

	// Global rate limit: 200 requests/min per IP
	server.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
		Requests: 200,
		Window:   time.Minute,
		Message:  "Too Many Requests",
	}))

	server.Use(middleware.Timeout(middleware.TimeoutConfig{Duration: 30 * time.Second}))

	// Double-submit cookie CSRF. Safe methods (GET/HEAD/OPTIONS/TRACE) are
	// skipped; the token cookie must be echoed back via X-CSRF-Token on all
	// unsafe requests. The Stripe webhook is server-to-server and cannot carry
	// a browser CSRF token, so it is exempt.
	csrf := middleware.CSRF(middleware.CSRFConfig{
		CookieName:     "_csrf",
		HeaderName:     "X-CSRF-Token",
		CookieSecure:   cfg.SecureCookie,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
		CookieMaxAge:   int((24 * time.Hour).Seconds()),
		SkipSafe:       true,
	})
	server.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/webhook/stripe" {
				next.ServeHTTP(w, r)
				return
			}
			csrf(next).ServeHTTP(w, r)
		})
	})

	// Health and readiness checks
	api := server.Group("/api")
	api.GET("/health", func(c *kern.Context) { _ = c.JSON(http.StatusOK, map[string]string{"status": "ok"}) })
	api.GET("/ready", func(c *kern.Context) {
		ctx := c.Context()
		if err := db.Ping(ctx); err != nil {
			_ = c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		_ = c.JSON(http.StatusOK, map[string]string{"status": "ready"})
	})

	// CSRF token for the SPA (safe GET — sets/rotates the _csrf cookie and
	// returns its value so the frontend can echo it via X-CSRF-Token).
	api.GET("/csrf", func(c *kern.Context) {
		token, _ := middleware.CSRFTokenFromContext(c.Context())
		_ = c.JSON(http.StatusOK, map[string]string{"csrf_token": token})
	})

	// Go runtime + process metrics for the observability stack
	registerMetrics(server)

	// Auth rate limit: 10 requests/min per IP (login, register, forgot-password)
	authRateLimit := middleware.RateLimiter(middleware.RateLimiterConfig{
		Requests: 10,
		Window:   time.Minute,
		Message:  "Too many attempts. Try again later.",
	})

	jwtAuth := shared.JWTAuth(a.TokenValidator)

	a.RegisterRoutes(server, jwtAuth, authRateLimit)

	server.Static("/", "./web/dist")

	slogger.Info("server started", "addr", cfg.Addr)
	if err := server.Run(cfg.Addr); err != nil {
		slogger.Error("server error", "error", err)
		os.Exit(1)
	}
}
