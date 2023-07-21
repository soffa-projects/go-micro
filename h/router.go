package h

import "github.com/fabriqs/go-micro/router"

func WithBind[T any](c router.Ctx, cb func(model T) (any, error)) (any, error) {
	m := new(T)
	if err := c.Bind(m); err != nil {
		return nil, err
	}
	return cb(*m)
}
