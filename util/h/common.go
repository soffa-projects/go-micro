package h

import (
	log "github.com/sirupsen/logrus"
	"reflect"
	"strconv"
	"strings"
)

func IsNil(v interface{}) bool {
	return v == nil
}

func IsNotNil(v interface{}) bool {
	return v != nil
}

func IsString(value interface{}) bool {
	_, ok := value.(string)
	return ok
}

func IsStrEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	if !IsString(v) {
		return false
	}
	var value string
	if IsPointer(v) {
		value = *v.(*string)
	} else {
		value = v.(string)
	}
	return len(strings.TrimSpace(value)) == 0
}

func IsPointer(arg interface{}) bool {
	argType := reflect.TypeOf(arg)
	return argType.Kind() == reflect.Ptr
}

func ToInt(input string) int {
	if input == "" {
		return 0
	}
	res, _ := strconv.Atoi(input)
	return res
}

func ToFloat(input string) float32 {
	if input == "" {
		return 0
	}
	res, err := strconv.ParseFloat(input, 32)
	if err != nil {
		log.Error(err)
		return 0
	}
	return float32(res)
}
