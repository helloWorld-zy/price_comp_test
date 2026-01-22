package http

import (
	"cruise-price-compare/internal/auth"
	"cruise-price-compare/internal/obs"

	"github.com/gin-gonic/gin"
)

// RouterConfig holds router configuration
type RouterConfig struct {
	Mode string // "debug", "release", "test"
}

// SetupRouter creates and configures the gin router
func SetupRouter(config RouterConfig, jwtService *auth.JWTService, logger *obs.Logger, metrics *obs.Metrics) *gin.Engine {
	if config.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else if config.Mode == "test" {
		gin.SetMode(gin.TestMode)
	}

	r := gin.New()

	// Add recovery middleware
	r.Use(gin.Recovery())

	// Add custom middlewares
	r.Use(obs.TraceMiddleware())
	r.Use(obs.RequestLoggerMiddleware(logger))
	r.Use(obs.MetricsMiddleware(metrics))
	r.Use(CORSMiddleware())
	r.Use(auth.NewUserContextMiddleware(jwtService).Handler())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"trace":  obs.GetTraceID(c),
		})
	})

	// Metrics endpoint
	r.GET("/metrics", func(c *gin.Context) {
		c.JSON(200, metrics.GetStats())
	})

	return r
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Trace-ID, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Trace-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RegisterRoutes registers all API routes
func RegisterRoutes(r *gin.Engine, handlers *Handlers) {
	// API v1 group
	v1 := r.Group("/api/v1")

	// Public routes (no auth required)
	public := v1.Group("")
	{
		public.POST("/auth/login", handlers.Auth.Login)
		public.POST("/auth/refresh", handlers.Auth.Refresh)
	}

	// Protected routes (auth required)
	protected := v1.Group("")
	protected.Use(auth.RequireAuth())
	{
		// Current user
		protected.GET("/auth/me", handlers.Auth.GetCurrentUser)
		protected.POST("/auth/logout", handlers.Auth.Logout)
		protected.PUT("/auth/password", handlers.Auth.ChangePassword)

		// Catalog - read (all authenticated users)
		protected.GET("/cruise-lines", handlers.Catalog.ListCruiseLines)
		protected.GET("/cruise-lines/:id", handlers.Catalog.GetCruiseLine)
		protected.GET("/ships", handlers.Catalog.ListShips)
		protected.GET("/ships/:id", handlers.Catalog.GetShip)
		protected.GET("/ships/:id/cabin-types", handlers.Catalog.ListCabinTypesByShip)
		protected.GET("/cabin-categories", handlers.Catalog.ListCabinCategories)
		protected.GET("/cabin-types", handlers.Catalog.ListCabinTypes)
		protected.GET("/cabin-types/:id", handlers.Catalog.GetCabinType)
		protected.GET("/sailings", handlers.Catalog.ListSailings)
		protected.GET("/sailings/:id", handlers.Catalog.GetSailing)
		protected.GET("/suppliers", handlers.Catalog.ListSuppliers)
		protected.GET("/suppliers/:id", handlers.Catalog.GetSupplier)

		// Quotes
		protected.GET("/quotes", handlers.Quote.ListQuotes)
		protected.GET("/quotes/:id", handlers.Quote.GetQuote)
		protected.POST("/quotes", handlers.Quote.CreateQuote)
		protected.PUT("/quotes/:id/void", handlers.Quote.VoidQuote)

		// Import
		protected.POST("/import/upload", handlers.Import.UploadFile)
		protected.GET("/import/jobs", handlers.Import.ListJobs)
		protected.GET("/import/jobs/:id", handlers.Import.GetJob)
		protected.POST("/import/jobs/:id/retry", handlers.Import.RetryJob)
	}

	// Admin routes
	admin := v1.Group("/admin")
	admin.Use(auth.RequireAuth(), auth.RequireAdmin())
	{
		// Catalog - write
		admin.POST("/cruise-lines", handlers.Catalog.CreateCruiseLine)
		admin.PUT("/cruise-lines/:id", handlers.Catalog.UpdateCruiseLine)
		admin.DELETE("/cruise-lines/:id", handlers.Catalog.DeleteCruiseLine)
		admin.POST("/ships", handlers.Catalog.CreateShip)
		admin.PUT("/ships/:id", handlers.Catalog.UpdateShip)
		admin.DELETE("/ships/:id", handlers.Catalog.DeleteShip)
		admin.POST("/cabin-categories", handlers.Catalog.CreateCabinCategory)
		admin.PUT("/cabin-categories/:id", handlers.Catalog.UpdateCabinCategory)
		admin.DELETE("/cabin-categories/:id", handlers.Catalog.DeleteCabinCategory)
		admin.POST("/cabin-types", handlers.Catalog.CreateCabinType)
		admin.PUT("/cabin-types/:id", handlers.Catalog.UpdateCabinType)
		admin.DELETE("/cabin-types/:id", handlers.Catalog.DeleteCabinType)
		admin.POST("/sailings", handlers.Catalog.CreateSailing)
		admin.PUT("/sailings/:id", handlers.Catalog.UpdateSailing)
		admin.DELETE("/sailings/:id", handlers.Catalog.DeleteSailing)
		admin.POST("/suppliers", handlers.Catalog.CreateSupplier)
		admin.PUT("/suppliers/:id", handlers.Catalog.UpdateSupplier)
		admin.DELETE("/suppliers/:id", handlers.Catalog.DeleteSupplier)
	}
}

// Handlers aggregates all HTTP handlers
type Handlers struct {
	Auth    *AuthHandler
	Catalog *CatalogHandler
	Quote   *QuoteHandler
	Import  *ImportHandler
}
