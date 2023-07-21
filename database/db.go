package database

import (
	"embed"
)

type Cfg struct {
	Url        string
	Migrations embed.FS
}

type BaseRepo struct {
	db DB
}

type DB interface {
	Transaction(func(tx DB) error) error
	Close()
	Create(target interface{}) error
	Save(target interface{}) error
	Find(target interface{}) error
	FindBy(target interface{}, where string, args ...interface{}) error
	Count(model interface{}) (int64, error)
	FindAll(target interface{}, orderBy string) error
	FindBySorted(target interface{}, orderBy string, where string, args ...interface{}) error
	DeleteAll(model interface{}) (int64, error)
	Ping() error
}
