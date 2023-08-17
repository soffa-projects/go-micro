package h

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"reflect"
)

func ToJsonStringPtr(input interface{}) (*string, error) {
	out, err := ToJsonString(input)
	return &out, err
}

func ToJsonString(input interface{}) (string, error) {
	if jsonBytes, err := json.Marshal(input); err != nil {
		log.Error("Error marshaling JSON:", err)
		return "", err
	} else {
		return string(jsonBytes), nil
	}
}

func Diff(original, updated interface{}) map[string]interface{} {
	originalValue := reflect.ValueOf(original)
	updatedValue := reflect.ValueOf(updated)

	// Check if the original and update values are pointers, and dereference them if they are
	if originalValue.Kind() == reflect.Ptr {
		originalValue = originalValue.Elem()
	}
	if updatedValue.Kind() == reflect.Ptr {
		updatedValue = updatedValue.Elem()
	}

	diff := make(map[string]interface{})

	for i := 0; i < originalValue.NumField(); i++ {
		originalField := originalValue.Field(i)
		updatedField := updatedValue.FieldByName(originalValue.Type().Field(i).Name)

		// If the update field is present, non-nil, and its value is different from the original field's value, add the field to the diff map
		if updatedField.IsValid() {
			if updatedField.Kind() == reflect.Ptr {
				if !updatedField.IsNil() && !reflect.DeepEqual(originalField.Interface(), updatedField.Elem().Interface()) {
					diff[ToSnakeCase(originalValue.Type().Field(i).Name)] = updatedField.Elem().Interface()
				}
			} else if !reflect.DeepEqual(originalField.Interface(), updatedField.Interface()) {
				diff[ToSnakeCase(originalValue.Type().Field(i).Name)] = updatedField.Interface()
			}
		}
	}

	return diff
}

func DeserializeJson(input string, out interface{}) error {
	return json.Unmarshal([]byte(input), out)
}
