package schema

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type Upload struct {
	Id       string  `json:"id,omitempty"`
	Name     string  `json:"name,omitempty"`
	Url      string  `json:"url,omitempty" validate:"required"`
	Mime     string  `json:"mime,omitempty"`
	Metadata string  `json:"metadata,omitempty"`
	Tags     *string `json:"tags,omitempty"`
}

//goland:noinspection GoMixedReceiverTypes
func (a Upload) ToString(input []*Upload) *string {
	if input == nil {
		return nil
	}
	var images []string
	for _, image := range input {
		images = append(images, image.Url)
	}
	result := strings.Join(images, ",")
	return &result
}

func (a *Upload) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type UploadList struct {
	Data []Upload `json:"-"`
}

func (f *UploadList) Value() (driver.Value, error) {
	return json.Marshal(f.Data)
}

func (f *UploadList) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Data)
}

func (f *UploadList) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	switch src := src.(type) {
	case []byte:
		return json.Unmarshal(src, &f.Data)
	case string:
		return json.Unmarshal([]byte(src), &f.Data)
	default:
		return fmt.Errorf("cannot scan type %T into MyData", src)
	}
}

func (f *UploadList) ToString() *string {
	var images []string
	for _, image := range f.Data {
		images = append(images, image.Url)
	}
	result := strings.Join(images, ",")
	return &result
}
