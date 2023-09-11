package micro

import (
	"github.com/fabriqs/go-micro/util/errors"
	"io/fs"
)

type DataSourceCfg struct {
	Production   bool
	Url          string
	Migrations   fs.FS
	TenantLoader TenantLoader
}

type DataSourceMigrations interface {
	Migrate()
	MigrateTenants(tenants []string)
	MigrateTenant(tenant string)
}

type DataSource interface {
	DataSourceMigrations
	//Transaction(func(tx DataSource) error) error
	Close()
	Save(ctx Ctx, target any) error
	Create(ctx Ctx, target any) error
	Ping(ctx Ctx) error
	Delete(Ctx, any, Query) (int64, error)
	Exists(Ctx, any, Query) (bool, error)
	First(Ctx, any, Query) error
	Find(Ctx, any, Query) error
	Count(Ctx, any, Query) (int64, error)
	Execute(Ctx, any, Query) (int64, error)
	Patch(ctx Ctx, model any, id string, data map[string]interface{}) (int64, error)
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
