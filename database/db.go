package database

import (
	"embed"
)

type Cfg struct {
	Url        string
	Migrations embed.FS
}

type DB interface {
	Transaction(func(tx DB) error) error
	Close()
	Create(target interface{}) error
	Save(target interface{}) error
	Find(target interface{}) error
	FindBy(target interface{}, where string, args ...interface{}) error
	FirstBy(target interface{}, where string, args ...interface{}) error
	Count(model interface{}) (int64, error)
	FindAll(target interface{}) error
	FindAllSorted(target interface{}, orderBy string) error
	FindBySorted(target interface{}, orderBy string, where string, args ...interface{}) error
	DeleteAll(model interface{}) (int64, error)
	Ping() error
}

type BaseRepo struct {
	DB    DB
	Model interface{}
}

func (r *BaseRepo) Count() (int64, error) {
	return r.DB.Count(r.Model)
}

func (r *BaseRepo) Save(data interface{}) error {
	return r.DB.Save(data)
}

func (r *BaseRepo) Create(data interface{}) error {
	return r.DB.Create(data)
}
