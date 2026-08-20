package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/mobentum/kern"
	"github.com/mobentum/kern/extensions/xlog"
	"github.com/mobentum/kern/extensions/xotel"
	"github.com/mobentum/kern/middleware"
	"github.com/mobentum/xdb"
	"github.com/sraj/addressbook/internal/app"
	"github.com/sraj/addressbook/internal/config"
	"github.com/sraj/addressbook/internal/shared"
	"github.com/sraj/addressbook/migrations"
)

func main() {
	// Initialize application logging before any startup operation so failures
	// are emitted in the same structured format as runtime logs.
	slogger := xlog.NewLogger(xlog.Config{Format: "json", Level: slog.LevelInfo})
	slog.SetDefault(slogger)

	// Load configuration before opening external resources.
	cfg, err := config.Load()
	if err != nil {
		slogger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if cfg.EnableOTLP {
		// Set up tracing and metrics for the lifetime of the server.
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
	}

	// Open the database and bring its schema up to date before constructing
	// application services or accepting requests.
	db, err := shared.NewDB(cfg.DatabaseDriver, cfg.DatabasePath)
	if err != nil {
		slogger.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	if err := migrations.MigrateUp(db, cfg.DatabaseDriver); err != nil {
		slogger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	application := app.New(db, cfg)

	server := initServer(slogger, cfg)
	registerRoutes(cfg, server, db, application)

	// Register the frontend last so API routes take precedence over the static
	// file handler for paths that share the root prefix.
	server.Static("/", "./web/dist")

	// Start serving only after every dependency, middleware, and route is ready.
	slogger.Info("server started", "addr", cfg.Addr)
	if err := server.Run(cfg.Addr); err != nil {
		slogger.Error("server error", "error", err)
		os.Exit(1)
	}
}

// initServer creates the HTTP application and installs global middleware.
func initServer(slogger *slog.Logger, cfg *config.Config) *kern.App {
	server := kern.New(kern.WithSlogLogger(slogger))

	// Recovery middleware to catch panics and return a 500 error
	server.Use(shared.Recovery())
	server.Use(middleware.RequestID())
	server.Use(xotel.Middleware(xotel.Config{ServiceName: "addressbook"}))
	server.Use(kern.Logger(kern.LoggerConfig{
		SLogger: slogger,
		Format:  "json",
		Fields:  map[string]interface{}{"log_type": "access"},
	}))

	// Security headers middleware
	server.Use(middleware.SecurityHeaders(middleware.SecurityHeadersConfig{
		StrictTransportSecurity: "max-age=31536000; includeSubDomains",
		ContentSecurityPolicy:   "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
		ReferrerPolicy:          "strict-origin-when-cross-origin",
		PermissionsPolicy:       "camera=(), microphone=(), geolocation=(), payment=()",
	}))

	// CORS runs before request guards so browser preflight requests can
	// complete without consuming rate-limit or CSRF checks.
	server.Use(kern.CORSWithConfig(kern.CORSConfig{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-CSRF-Token"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
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

	return server
}

// registerRoutes keeps operational endpoints ahead of application routes.
func registerRoutes(cfg *config.Config, server *kern.App, db *xdb.DB, application *app.App) {
	// Initialize OTLP-related middleware or handlers here if needed.
	registerOperationalRoutes(cfg, server, db)
	registerApplicationRoutes(server, application)
}

func registerOperationalRoutes(cfg *config.Config, server *kern.App, db *xdb.DB) {
	// Health and readiness endpoints are used by the process manager and load
	// balancer, so they are registered independently of user authentication.
	api := server.Group("/api")
	api.GET("/health", func(c *kern.Context) {
		_ = c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	api.GET("/ready", func(c *kern.Context) {
		ctx := c.Context()
		if err := db.Ping(ctx); err != nil {
			_ = c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		_ = c.JSON(http.StatusOK, map[string]string{"status": "ready"})
	})

	if cfg.EnableOTLP {
		// Initialize OTLP-related middleware or handlers here if needed.
		// Expose Go runtime and process metrics for the observability stack.
		registerMetrics(server)
	}
}

func registerApplicationRoutes(server *kern.App, application *app.App) {
	// Authentication routes get a stricter limit than the global request limit.
	authRateLimit := middleware.RateLimiter(middleware.RateLimiterConfig{
		Requests: 10,
		Window:   time.Minute,
		Message:  "Too many attempts. Try again later.",
	})

	jwtAuth := shared.JWTAuth(application.TokenValidator)

	api := server.Group("/api")
	// The SPA calls this safe endpoint to obtain the CSRF token cookie and the
	// token value it must echo in X-CSRF-Token for unsafe requests.
	api.GET("/v1/csrf", func(c *kern.Context) {
		token, _ := middleware.CSRFTokenFromContext(c.Context())
		_ = c.JSON(http.StatusOK, map[string]string{"csrf_token": token})
	})

	application.RegisterRoutes(server, jwtAuth, authRateLimit)
}
