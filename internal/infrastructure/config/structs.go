package Config

import "time"

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
	Port    string
	Timeout time.Duration
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

type Config struct {
	Environment Environment
	Database    DatabaseConfig
	Server      ServerConfig
	JWT         JWTConfig
	RateLimit   RateLimitConfig
	LogLevel    string
}

type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
)
