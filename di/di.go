package di

import (
	"github.com/fabriqs/go-micro/micro"
	"go.uber.org/dig"
)

var c = dig.New()

func Provide(provider interface{}) error {
	return c.Provide(provider)
}

func Overwrite(decorator interface{}) error {
	return c.Decorate(decorator)
}

func Inject(function interface{}) error {
	return c.Invoke(func(env *micro.Env) error {
		_, err := micro.Invoke(env, func() (interface{}, error) {
			err := c.Invoke(function)
			return nil, err
		})
		return err
	})
}

func Reset() {
	c = dig.New()
}
