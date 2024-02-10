package h

import (
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"reflect"
	"strconv"
	"strings"
)

func IsEmpty(v interface{}) bool {
	return funk.IsEmpty(v)
}

func IsNotEmpty(v interface{}) bool {
	return !funk.IsEmpty(v)
}

func IsString(value interface{}) bool {
	_, ok := value.(string)
	return ok
}

func IsStrEmpty(v interface{}) bool {
	if IsEmpty(v) {
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

func UnwrapStr(arg interface{}) string {
	if IsPointer(arg) {
		idValue := reflect.ValueOf(arg)
		if idValue.IsNil() {
			// Handle the nil pointer case appropriately
			return ""
		} else {
			return idValue.Elem().Interface().(string)
		}
	} else {
		return arg.(string)
	}
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
