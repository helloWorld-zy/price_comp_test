package app

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"cruise-price-compare/internal/auth"
	"cruise-price-compare/internal/llm"
	"cruise-price-compare/internal/obs"
	"cruise-price-compare/internal/repo"
	"cruise-price-compare/internal/service"
	httpTransport "cruise-price-compare/internal/transport/http"
)

// Config holds application configuration
type Config struct {
	// Server
	ServerHost string
	ServerPort int
	ServerMode string // debug, release, test

	// Database
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	// JWT
	JWTSecret          string
	JWTAccessTokenTTL  time.Duration
	JWTRefreshTokenTTL time.Duration

	// Logging
	LogLevel  string
	LogFormat string

	// Import
	UploadDir   string
	OllamaURL   string
	OllamaModel string
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() *Config {
	return &Config{
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort: getEnvInt("SERVER_PORT", 8080),
		ServerMode: getEnv("GIN_MODE", "debug"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvInt("DB_PORT", 3306),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "cruise_price"),

		JWTSecret:          getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
		JWTAccessTokenTTL:  getEnvDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
		JWTRefreshTokenTTL: getEnvDuration("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour),

		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),

		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
		OllamaURL:   getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel: getEnv("OLLAMA_MODEL", "llama2"),
	}
}

// Container holds all application dependencies
type Container struct {
	Config  *Config
	Logger  *obs.Logger
	Metrics *obs.Metrics
	DB      *repo.DB

	// Repositories
	UserRepo          *repo.UserRepository
	CruiseLineRepo    *repo.CruiseLineRepository
	ShipRepo          *repo.ShipRepository
	CabinCategoryRepo *repo.CabinCategoryRepository
	CabinTypeRepo     *repo.CabinTypeRepository
	SailingRepo       *repo.SailingRepository
	SupplierRepo      *repo.SupplierRepository
	PriceQuoteRepo    *repo.PriceQuoteRepository
	ImportJobRepo     *repo.ImportJobRepository
	AuditLogRepo      *repo.AuditLogRepository

	// Services
	JWTService         *auth.JWTService
	PasswordService    *auth.PasswordService
	AuthService        *auth.AuthService
	AuditService       *obs.AuditService
	CatalogService     *service.CatalogService
	QuoteService       *service.QuoteService
	ImportJobService   *service.ImportJobService
	FileStorageService *service.FileStorageService

	// HTTP Handlers
	Handlers *httpTransport.Handlers
}

// NewContainer creates a new dependency injection container
func NewContainer(config *Config) (*Container, error) {
	c := &Container{Config: config}

	// Initialize logger
	c.Logger = obs.NewLogger(obs.LogConfig{
		Level:  obs.LogLevel(config.LogLevel),
		Format: config.LogFormat,
	})
	obs.SetDefault(c.Logger)

	// Initialize metrics
	c.Metrics = obs.NewMetrics()

	// Initialize database
	db, err := repo.NewDB(repo.Config{
		Host:     config.DBHost,
		Port:     config.DBPort,
		User:     config.DBUser,
		Password: config.DBPassword,
		Database: config.DBName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	c.DB = db

	// Initialize repositories
	c.UserRepo = repo.NewUserRepository(db)
	c.CruiseLineRepo = repo.NewCruiseLineRepository(db)
	c.ShipRepo = repo.NewShipRepository(db)
	c.CabinCategoryRepo = repo.NewCabinCategoryRepository(db)
	c.CabinTypeRepo = repo.NewCabinTypeRepository(db)
	c.SailingRepo = repo.NewSailingRepository(db)
	c.SupplierRepo = repo.NewSupplierRepository(db)
	c.PriceQuoteRepo = repo.NewPriceQuoteRepository(db)
	c.ImportJobRepo = repo.NewImportJobRepository(db)
	c.AuditLogRepo = repo.NewAuditLogRepository(db)

	// Initialize auth services
	c.JWTService = auth.NewJWTService(auth.JWTConfig{
		SecretKey:       config.JWTSecret,
		AccessTokenTTL:  config.JWTAccessTokenTTL,
		RefreshTokenTTL: config.JWTRefreshTokenTTL,
	})
	c.PasswordService = auth.NewPasswordService(nil)
	c.AuthService = auth.NewAuthService(c.UserRepo, c.JWTService, c.PasswordService)

	// Initialize services
	c.AuditService = obs.NewAuditService(c.AuditLogRepo, c.Logger)
	c.CatalogService = service.NewCatalogService(
		c.CruiseLineRepo, c.ShipRepo, c.CabinCategoryRepo, c.CabinTypeRepo,
		c.SailingRepo, c.SupplierRepo, c.AuditService, c.Logger,
	)

	// Initialize quote service
	c.QuoteService = service.NewQuoteService(
		c.PriceQuoteRepo,
		c.SailingRepo,
		c.CabinTypeRepo,
		c.SupplierRepo,
		c.AuditService,
	)

	// Initialize file storage and import services
	c.FileStorageService = service.NewFileStorageService(config.UploadDir)
	ollamaClient := llm.NewOllamaClient(config.OllamaURL, config.OllamaModel)
	dataMatcher := service.NewDataMatcher(
		c.ShipRepo,
		c.SailingRepo,
		c.CabinTypeRepo,
		c.CruiseLineRepo,
	)
	c.ImportJobService = service.NewImportJobService(
		c.ImportJobRepo,
		c.FileStorageService,
		ollamaClient,
		dataMatcher,
		c.QuoteService,
		c.AuditService,
	)

	// Initialize HTTP handlers
	c.Handlers = &httpTransport.Handlers{
		Auth:    httpTransport.NewAuthHandler(c.AuthService),
		Catalog: httpTransport.NewCatalogHandler(c.CatalogService),
		Quote:   httpTransport.NewQuoteHandler(c.QuoteService),
		Import:  httpTransport.NewImportHandler(c.ImportJobService),
	}

	c.Logger.Info("application container initialized")
	return c, nil
}

// Close closes all resources
func (c *Container) Close() error {
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}
	c.Logger.Info("application container closed")
	return nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultValue
}
