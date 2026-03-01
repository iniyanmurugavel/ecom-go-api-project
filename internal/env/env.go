// Package env provides helpers for reading config from environment (e.g. after godotenv.Load(".env")).
package env

import (
	"os"
	"strconv"
	"time"
)

// GetString returns os.Getenv(key), or fallback if unset/empty. Used for HTTP_ADDR, DB_*, etc.
func GetString(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// GetInt returns the integer value of os.Getenv(key), or fallback if unset/empty/invalid.
func GetInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return fallback
}

// GetInt32 returns the int32 value of os.Getenv(key), or fallback if unset/empty/invalid.
func GetInt32(key string, fallback int32) int32 {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.ParseInt(val, 10, 32); err == nil {
			return int32(n)
		}
	}
	return fallback
}

// GetDuration returns the duration from os.Getenv(key). Supports "1h", "30m", "24h", or integer seconds.
// Falls back to fallback if unset/empty/invalid.
func GetDuration(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return fallback
}