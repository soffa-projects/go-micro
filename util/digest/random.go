package digest

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/google/uuid"
)

func GenerateClientId() string {
	return uuid.New().String()
}

func GenerateClientSecret(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
