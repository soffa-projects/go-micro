package h

import "encoding/base64"

func EncoreBase64(message string) string {
	messageBytes := []byte(message)
	encoded := base64.StdEncoding.EncodeToString(messageBytes)
	return encoded
}
