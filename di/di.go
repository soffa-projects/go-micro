package di

import "go.uber.org/dig"

var c = dig.New()

func Provide(provider interface{}) error {
	return c.Provide(provider)
}

func Overwrite(decorator interface{}) error {
	return c.Decorate(decorator)
}

func Inject(function interface{}) error {
	return c.Invoke(function)
}

func Reset() {
	c = dig.New()
}
