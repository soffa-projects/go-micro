package h

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func GetEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}

func GetEnvOrDefault(key string, defaultValue string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
	return defaultValue
}

func RequireEnv(key string) string {
	return RequireEnvIf(true, key)
}

func RequireEnvIf(test bool, key string) string {
	value := os.Getenv(key)
	if test && value == "" {
		log.Fatalf("Missing env variable: %s", key)
	}
	return value
}
