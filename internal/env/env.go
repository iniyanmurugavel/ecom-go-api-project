// Package env provides helpers for reading config from environment (e.g. after godotenv.Load(".env")).
package env

import (
	"os"
	"strconv"
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