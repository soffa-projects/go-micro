package di

import (
	log "github.com/sirupsen/logrus"
	"reflect"
)

/*
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
*/

type Component interface {
}

var registry = make(map[string]any)

func Register(name string, provider interface{}) {
	registry[name] = provider
}

func Resolve[T Component](typ T) *T {
	rtype := reflect.TypeOf(typ)
	for _, component := range registry {
		cr := reflect.TypeOf(component)
		if cr == rtype {
			return component.(*T)
		}
		if cr.Kind() == reflect.Ptr && cr.Elem() == rtype {
			return component.(*T)
		}
	}
	log.Fatalf("failed to resolve component %s", typ)
	return nil
}

func ResolveByName[T interface{}](name string) T {
	if component, ok := registry[name]; ok {
		return component.(T)
	}
	log.Fatalf("failed to resolve component %s", name)
	panic("failed to resolve component")
}

func Clear() {
	registry = make(map[string]any)
}
