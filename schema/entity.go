package schema

type IdModel struct {
	Id *string `param:"id" json:"id" validate:"required"`
}

type EntityList[T any] struct {
	Data  []T `json:"data"`
	Page  int `json:"page,omitempty"`
	Total int `json:"total,omitempty"`
}
