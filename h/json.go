package h

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

func ToJsonString(input interface{}) (string, error) {
	if jsonBytes, err := json.Marshal(input); err != nil {
		log.Error("Error marshaling JSON:", err)
		return "", err
	} else {
		return string(jsonBytes), nil
	}
}
