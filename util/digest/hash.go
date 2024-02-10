package digest

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
)

func Sha256(data string) string {
	hasher := sha256.New()
	hasher.Write([]byte(data))
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func EncodeToBase64String(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func DecodeBase64(input string) string {
	data, _ := base64.StdEncoding.DecodeString(input)
	return string(data)
}
