package h

import (
	"math/rand"
	"strings"
	"time"
)

func ToSnakeCase(str string) string {
	var output []rune
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			output = append(output, '_')
		}
		output = append(output, r)
	}
	return strings.ToLower(string(output))
}

func PtrStr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func Str(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func Ptr[T any](value T) *T {
	return &value
}

func RandomString(length int) string {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	characters := "abcdefghijklmnopqrstuvwxyz0123456789"
	result := ""
	for i := 0; i < length; i++ {
		randomIndex := random.Intn(len(characters))
		result += string(characters[randomIndex])
	}
	return result
}
