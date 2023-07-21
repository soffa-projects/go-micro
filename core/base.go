package core

import (
	"github.com/fabriqs/go-micro/database"
	"github.com/fabriqs/go-micro/policy"
	"github.com/fabriqs/go-micro/router"
)

type Feature struct {
	Name string
	Init func(router router.R) error
}

type App struct {
	Name     string
	Features []Feature
	Router   router.R
	//Env      *Env
}

type Env struct {
	context map[string]interface{}
	DB      database.DB
	Policy  policy.Manager
}

type AppCfg struct {
	Name     string
	Features []Feature
	Router   router.R
	DB       database.DB
}

func NewEnv(env *Env) *Env {
	return &Env{
		context: map[string]interface{}{},
		DB:      env.DB,
		Policy:  env.Policy,
	}
}

func (e *Env) Set(key string, value interface{}) {
	e.context[key] = value
}

func (e *Env) Get(key string) interface{} {
	if val, ok := e.context[key]; ok {
		return val
	}
	return nil
}
