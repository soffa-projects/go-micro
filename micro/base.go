package micro

import "github.com/fabriqs/go-micro/di"

type Feature struct {
	Name string
	//@deprecated
	Init func(app *App) (di.Component, error)
}

type App struct {
	Name          string
	Features      []Feature
	Router        Router
	Scheduler     Scheduler
	TokenProvider TokenProvider
	Notifier      NotificationService
	Mailer        Mailer
	Env           *Env
	//components    []Component
}

type Env struct {
	Ctx
	Conf       interface{}
	DataSource DataSource
}

type AppCfg struct {
	Name     string
	Features []Feature
	Router   Router
	DB       DataSource
}

func (e *Env) DB() DataSource {
	return e.DataSource
}

func (e *Env) Config() interface{} {
	return e.Conf
}

func (e *Env) Tx(callback func(tx DataSource) (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := e.DB().Transaction(func(tx DataSource) error {
		result0, err0 := callback(tx)
		result = result0
		return err0
	})
	return result, err
}
