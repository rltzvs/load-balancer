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
	Logging    LoggingConfig
	Balancer   BalancerConfig
	RateLimits RateLimitConfig
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
}

type RateLimitConfig struct {
	DefaultCapacity int
	DefaultRate     int
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
		},
		RateLimits: RateLimitConfig{
			DefaultCapacity: getIntEnv("RATE_LIMIT_DEFAULT_CAPACITY", 100),
			DefaultRate:     getIntEnv("RATE_LIMIT_DEFAULT_RATE", 10),
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
	return nil
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
