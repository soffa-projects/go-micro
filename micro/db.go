package micro

import (
	"embed"
	"github.com/fabriqs/go-micro/util/errors"
)

type DataSourceCfg struct {
	Url        string
	Migrations embed.FS
}

type DataSource interface {
	Transaction(func(tx DataSource) error) error
	Close()
	Create(target any) error
	Save(target any) error
	Ping() error
	Delete(any, Query) (int64, error)
	Exists(any, Query) (bool, error)
	First(any, Query) error
	Find(any, Query) error
	Count(any, Query) (int64, error)
	Execute(any, Query) (int64, error)
	Patch(model any, id string, data map[string]interface{}) (int64, error)
	Migrate()
}

var ErrRecordNotFound = errors.Functional("record not found")

type Query struct {
	Model  any
	Raw    string
	W      string
	Sort   string
	Args   []any
	Select string
	Offset int64
	Limit  int64
}
