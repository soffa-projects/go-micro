package h

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"reflect"
	"strings"
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

func ToMap(input interface{}) (map[string]interface{}, error) {
	data, err := ToJsonString(input)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal([]byte(data), &result)
	return result, err
}

func ToJsonBytes(input interface{}) ([]byte, error) {
	if jsonBytes, err := json.Marshal(input); err != nil {
		log.Error("Error marshaling JSON:", err)
		return nil, err
	} else {
		return jsonBytes, nil
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

func DeserializeJsonBytes(input []byte, out interface{}) error {
	return json.Unmarshal(input, out)
}

func DeserializeJsonUri(input string, out interface{}) error {
	if strings.HasPrefix(input, "http") {
		resp, err := http.Get(input)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return json.NewDecoder(resp.Body).Decode(out)
	} else {
		// read file from disk
		data, err := os.ReadFile(input)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, out)
	}

}
