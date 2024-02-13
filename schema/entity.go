package schema

type IdModel struct {
	Id *string `param:"id" json:"id" validate:"required"`
}

type EntityList[T any] struct {
	Data  []T `json:"data"`
	Page  int `json:"page,omitempty"`
	Total int `json:"total,omitempty"`
}

type FilterInput struct {
	Where string        `json:"where" query:"where" validate:"required"`
	Args  []interface{} `json:"args" query:"args"`
	Page  int           `json:"page" query:"page"`
	Sort  string        `json:"sort" query:"sort"`
	Count int           `json:"count" query:"count"`
}

type PagingInput struct {
	Page  int    `json:"page" query:"page"`
	Sort  string `json:"sort" query:"sort"`
	Count int    `json:"count" query:"count"`
}
