package micro

import (
	gerror "errors"
	"github.com/soffa-projects/go-micro/util/errors"
	"io/fs"
)

/*type DataSourceCfg struct {
	Production   bool
	Url          string
	Migrations   fs.FS
	TenantLoader TenantLoader
}*/

const DefaultMigrationsTable = "z_migrations"

type DataSourceMigrations interface {
	Migrate(fs fs.FS, location string, migrationsTable string)
}

type EntityHooks interface {
	PreCreate() error
}

type DataSource interface {
	IsPostgres() bool
	Tenant() string
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
	FindAll(any) error
	Count(any, Query) (int64, error)
	Execute(any, Query) (int64, error)
	Raw(Query) (int64, error)
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

type SimpleRepo[T any] struct {
	db     DataSource
	entity T
}

func NewSimpleRepo[T any](db DataSource, entity T) *SimpleRepo[T] {
	return &SimpleRepo[T]{db: db, entity: entity}
}

func (r *SimpleRepo[T]) Count(q ...Query) (int64, error) {
	var model T
	if len(q) == 0 {
		return r.db.Count(model, Query{})
	}
	return r.db.Count(model, q[0])
}

func (r *SimpleRepo[T]) FirstBy(q string, args ...any) (*T, error) {
	var model *T
	err := r.db.First(&model, Query{W: q, Args: args})
	if err != nil && gerror.Is(err, ErrRecordNotFound) {
		return nil, nil
	}
	return model, err
}
