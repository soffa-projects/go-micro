package micro

import (
	"github.com/soffa-projects/go-micro/util/errors"
	"io/fs"
)

/*type DataSourceCfg struct {
	Production   bool
	Url          string
	Migrations   fs.FS
	TenantLoader TenantLoader
}*/

type DataSourceMigrations interface {
	Migrate(fs fs.FS, location string)
}

type DataSource interface {
	DataSourceMigrations
	Transaction(func(tx DataSource) error) error
	Close()
	Save(target any) error
	Create(target any) error
	Ping() error
	Delete(any, Query) (int64, error)
	Exists(any, Query) (bool, error)
	First(any, Query) error
	Find(any, Query) error
	Count(any, Query) (int64, error)
	Execute(any, Query) (int64, error)
	Patch(model any, id string, data map[string]interface{}) (int64, error)
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
