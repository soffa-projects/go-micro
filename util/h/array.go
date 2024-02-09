package h

import "strings"

func Contains(arr []string, str string) bool {
	for _, v := range arr {
		if strings.TrimSpace(v) == strings.TrimSpace(str) {
			return true
		}
	}
	return false
}
