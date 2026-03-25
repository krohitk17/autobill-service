package Config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

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
		var errorMsg strings.Builder
		errorMsg.WriteString("Configuration errors:\n")
		for _, err := range errors {
			fmt.Fprintf(&errorMsg, "  - %s\n", err)
		}
		panic(errorMsg.String())
	}

	timeout := optionalDurationEnvVar("TIMEOUT", 10*time.Second)
	env := Environment(optionalEnvVar("ENV", "development"))
	sslMode := optionalEnvVar("DATABASE_SSL_MODE", "disable")
	logLevel := optionalEnvVar("LOG_LEVEL", "info")
	maxOpenConns := optionalIntEnvVar("DATABASE_MAX_OPEN_CONNS", 25)
	maxIdleConns := optionalIntEnvVar("DATABASE_MAX_IDLE_CONNS", 5)
	connMaxLifetime := optionalDurationEnvVar("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute)
	jwtExpiration := optionalDurationEnvVar("JWT_EXPIRATION", 15*time.Minute)
	refreshTokenExpiration := optionalDurationEnvVar("REFRESH_TOKEN_EXPIRATION", 7*24*time.Hour)
	rateLimitMax := optionalIntEnvVar("RATE_LIMIT_MAX", 100)
	rateLimitWindow := optionalDurationEnvVar("RATE_LIMIT_WINDOW", 1*time.Minute)

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
			Port:    port,
			Timeout: timeout,
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
		LogLevel: logLevel,
	}
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == Development
}

func LoadTestConfig() Config {
	_ = godotenv.Load("/Users/rohit/Programs/AutoBill/autobill-service/.env.test")

	var errors []string

	dbHost := requiredEnvVar("TEST_DATABASE_HOST", &errors)
	dbName := requiredEnvVar("TEST_DATABASE_NAME", &errors)
	dbUser := requiredEnvVar("TEST_DATABASE_USER", &errors)
	dbPass := requiredEnvVar("TEST_DATABASE_PASSWORD", &errors)
	dbPort := requiredEnvVar("TEST_DATABASE_PORT", &errors)

	if len(errors) > 0 {
		var errorMsg strings.Builder
		errorMsg.WriteString("Configuration errors:\n")
		for _, err := range errors {
			fmt.Fprintf(&errorMsg, "  - %s\n", err)
		}
		panic(errorMsg.String())
	}

	return Config{
		Database: DatabaseConfig{
			Host:         dbHost,
			Port:         dbPort,
			User:         dbUser,
			Password:     dbPass,
			Name:         dbName,
			SSLMode:      "disable",
			MaxOpenConns: 10,
			MaxIdleConns: 2,
			MaxLifetime:  5 * time.Minute,
		},
	}
}
