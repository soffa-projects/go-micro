package h

type Map map[string]interface{}

type MapS map[string]string

func MapLookup[T interface{}](values map[string]T, keys ...string) (T, bool) {
	for _, key := range keys {
		if value, ok := values[key]; ok && IsNotEmpty(value) && !IsStrEmpty(value) {
			return value, true
		}
	}
	var zero T // zero value of type T
	return zero, false
}
