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

func RequireEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Missing env varaible: %s", key)
	}
	return value
}
