package di

import "go.uber.org/dig"

var c = dig.New()

func Provide(constructor interface{}) error {
	return c.Provide(constructor)
}

func Inject(function interface{}) error {
	return c.Invoke(function)
}

func Reset() {
	c = dig.New()
}
