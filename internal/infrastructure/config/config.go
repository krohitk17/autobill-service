package Config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type JWTConfig struct {
	Secret                 string
	Expiration             time.Duration
	RefreshTokenExpiration time.Duration
}

type RateLimitConfig struct {
	MaxRequests int
	Window      time.Duration
}

type CORSConfig struct {
	AllowOrigins     string
	AllowMethods     string
	AllowHeaders     string
	AllowCredentials bool
}

type Config struct {
	Environment Environment
	Database    DatabaseConfig
	Server      ServerConfig
	JWT         JWTConfig
	RateLimit   RateLimitConfig
	CORS        CORSConfig
	LogLevel    string
	DBUrl       string
}

func requiredEnvVar(key string, errors *[]string) string {
	value := os.Getenv(key)
	if value == "" {
		*errors = append(*errors, fmt.Sprintf("missing required environment variable: %s", key))
	}
	return value
}

func optionalEnvVar(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func optionalIntEnvVar(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func optionalDurationEnvVar(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}

func Load() Config {
	_ = godotenv.Load()

	var errors []string

	dbHost := requiredEnvVar("DATABASE_HOST", &errors)
	dbName := requiredEnvVar("DATABASE_NAME", &errors)
	dbUser := requiredEnvVar("DATABASE_USER", &errors)
	dbPass := requiredEnvVar("DATABASE_PASSWORD", &errors)
	dbPort := requiredEnvVar("DATABASE_PORT", &errors)
	jwtSecret := requiredEnvVar("JWT_SECRET", &errors)
	port := requiredEnvVar("PORT", &errors)

	if len(errors) > 0 {
		errorMsg := "Configuration errors:\n"
		for _, err := range errors {
			errorMsg += fmt.Sprintf("  - %s\n", err)
		}
		panic(errorMsg)
	}

	env := Environment(optionalEnvVar("ENV", "development"))
	sslMode := optionalEnvVar("DATABASE_SSL_MODE", "disable")
	logLevel := optionalEnvVar("LOG_LEVEL", "info")
	maxOpenConns := optionalIntEnvVar("DATABASE_MAX_OPEN_CONNS", 25)
	maxIdleConns := optionalIntEnvVar("DATABASE_MAX_IDLE_CONNS", 5)
	connMaxLifetime := optionalDurationEnvVar("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute)
	readTimeout := optionalDurationEnvVar("SERVER_READ_TIMEOUT", 10*time.Second)
	writeTimeout := optionalDurationEnvVar("SERVER_WRITE_TIMEOUT", 10*time.Second)
	jwtExpiration := optionalDurationEnvVar("JWT_EXPIRATION", 15*time.Minute)
	refreshTokenExpiration := optionalDurationEnvVar("REFRESH_TOKEN_EXPIRATION", 7*24*time.Hour)
	rateLimitMax := optionalIntEnvVar("RATE_LIMIT_MAX", 100)
	rateLimitWindow := optionalDurationEnvVar("RATE_LIMIT_WINDOW", 1*time.Minute)

	corsOrigins := optionalEnvVar("CORS_ALLOW_ORIGINS", "*")
	corsMethods := optionalEnvVar("CORS_ALLOW_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	corsHeaders := optionalEnvVar("CORS_ALLOW_HEADERS", "Origin,Content-Type,Accept,Authorization,X-Request-ID")
	corsCredentials := optionalEnvVar("CORS_ALLOW_CREDENTIALS", "false") == "true"

	return Config{
		Environment: env,
		Database: DatabaseConfig{
			Host:         dbHost,
			Port:         dbPort,
			User:         dbUser,
			Password:     dbPass,
			Name:         dbName,
			SSLMode:      sslMode,
			MaxOpenConns: maxOpenConns,
			MaxIdleConns: maxIdleConns,
			MaxLifetime:  connMaxLifetime,
		},
		Server: ServerConfig{
			Port:         port,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
		JWT: JWTConfig{
			Secret:                 jwtSecret,
			Expiration:             jwtExpiration,
			RefreshTokenExpiration: refreshTokenExpiration,
		},
		RateLimit: RateLimitConfig{
			MaxRequests: rateLimitMax,
			Window:      rateLimitWindow,
		},
		CORS: CORSConfig{
			AllowOrigins:     corsOrigins,
			AllowMethods:     corsMethods,
			AllowHeaders:     corsHeaders,
			AllowCredentials: corsCredentials,
		},
		LogLevel: logLevel,
		DBUrl: fmt.Sprintf(
			"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
			dbHost, dbPort, dbUser, dbName, dbPass, sslMode,
		),
	}
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == Development
}
