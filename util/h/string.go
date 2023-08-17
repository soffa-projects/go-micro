package h

import "strings"

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
