package h

import (
	"github.com/jinzhu/copier"
	"reflect"
)

func CopyAllFields(dst any, src any, ignoreEmpty bool) error {
	return copier.CopyWithOption(dst, src, copier.Option{
		IgnoreEmpty: ignoreEmpty,
		DeepCopy:    true,
	})
}

func CopyFields(dst any, src any, excludedFields ...string) {
	srcValue := reflect.ValueOf(src)
	dstValue := reflect.ValueOf(dst).Elem()
	srcType := srcValue.Type()

	if srcType.Kind() == reflect.Ptr {
		// If src is a pointer, get the element it points to
		srcValue = srcValue.Elem()
		srcType = srcValue.Type()
	}

	for i := 0; i < srcValue.NumField(); i++ {
		srcField := srcValue.Field(i)
		dstField := dstValue.Field(i)
		fieldName := srcType.Field(i).Name

		// Check if the field is exportable and not in the list of excluded fields
		if dstField.CanSet() && !contains(excludedFields, fieldName) {
			dstField.Set(srcField)
		}
	}
}

func contains(fields []string, fieldName string) bool {
	for _, field := range fields {
		if field == fieldName {
			return true
		}
	}
	return false
}
