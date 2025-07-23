package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Logging    LoggingConfig
	Balancer   BalancerConfig
	RateLimits RateLimitDefaults
}

type ServerConfig struct {
	Port            string
	ShutdownTimeout time.Duration
}

type LoggingConfig struct {
	Level string
}

type BalancerConfig struct {
	Upstreams           []string
	HealthCheckInterval time.Duration
	Algorithm           string // round-robin, random, etc.
}

type RateLimitDefaults struct {
	DefaultCapacity int
	DefaultRate     int
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	SSLMode  string
}

func Load() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, err
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", "8080"),
			ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		Balancer: BalancerConfig{
			Upstreams:           parseUpstreams(getEnv("UPSTREAMS", "")),
			HealthCheckInterval: getDurationEnv("HEALTH_CHECK_INTERVAL", 5*time.Second),
			Algorithm:           getEnv("BALANCER_ALGORITHM", "round-robin"),
		},
		RateLimits: RateLimitDefaults{
			DefaultCapacity: getIntEnv("RATE_LIMIT_DEFAULT_CAPACITY", 100),
			DefaultRate:     getIntEnv("RATE_LIMIT_DEFAULT_RATE", 10),
		},
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			Database: getEnv("POSTGRES_DB", "balance_db"),
			Username: getEnv("POSTGRES_USER", "balancer"),
			Password: getEnv("POSTGRES_PASSWORD", "password"),
			SSLMode:  getEnv("POSTGRES_SSL_MODE", "disable"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if len(c.Balancer.Upstreams) == 0 {
		return fmt.Errorf("UPSTREAMS must be provided (comma-separated URLs)")
	}
	if c.Database.Host == "" || c.Database.Username == "" || c.Database.Password == "" {
		return fmt.Errorf("incomplete PostgreSQL configuration")
	}
	return nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
	)
}

func getEnv(key string, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getIntEnv(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}

func parseUpstreams(value string) []string {
	raw := strings.Split(value, ",")
	clean := make([]string, 0, len(raw))
	for _, b := range raw {
		trimmed := strings.TrimSpace(b)
		if trimmed != "" {
			clean = append(clean, trimmed)
		}
	}
	return clean
}
