package adapters

import (
  _ "github.com/jackc/pgx/v5"
  "github.com/onrik/gorm-logrus"
  "github.com/pressly/goose/v3"
  log "github.com/sirupsen/logrus"
  "github.com/soffa-projects/go-micro/micro"
  "gorm.io/driver/postgres"
  "gorm.io/driver/sqlite"
  "gorm.io/gorm"
  "io/fs"
  "strings"
)

type adapter struct {
	micro.DataSource
	//tenants map[string]*gorm.DB
	internal *gorm.DB
	tenantId string
	url      string
	//config  *micro.DataSourceCfg
}

func (a adapter) IsPostgres() bool {
	return strings.HasPrefix(a.url, "postgres")
}

func (a adapter) Tenant() string {
	return a.tenantId
}

func (a adapter) Create(model interface{}) error {
	if model, ok := model.(micro.EntityHooks); ok {
		if err := model.PreCreate(); err != nil {
			return err
		}
	}
	return a.internal.Create(model).Error
}

func (a adapter) Save(model interface{}) error {
	return a.internal.Save(model).Error
}

func (a adapter) Exists(model any, q micro.Query) (bool, error) {
	var count int64
	res := a.buildQuery(model, q).Count(&count)
	return count > 0, res.Error
}

func (a adapter) Find(target any, q micro.Query) error {
	res := a.buildQuery(target, q).Find(target)
	return res.Error
}

func (a adapter) FindAll(target any) error {
	res := a.buildQuery(target, micro.Query{}).Find(target)
	return res.Error
}

func (a adapter) First(model any, q micro.Query) error {
	res := a.buildQuery(model, q).First(model)
	if res.Error == gorm.ErrRecordNotFound {
		return micro.ErrRecordNotFound
	}
	if res.RowsAffected == 0 {
		return micro.ErrRecordNotFound
	}
	return res.Error
}

func (a adapter) Count(model any, q micro.Query) (int64, error) {
	var count int64
	res := a.buildQuery(model, q).Count(&count)
	return count, res.Error
}

func (a adapter) Delete(model any, q micro.Query) (int64, error) {
	res := a.buildQuery(model, q).Delete(model)
	return res.RowsAffected, res.Error
}

func (a adapter) Execute(model any, q micro.Query) (int64, error) {
	res := a.internal.Raw(q.Raw).Scan(model)
	return res.RowsAffected, res.Error
}

func (a adapter) Raw(q micro.Query) (int64, error) {
	res := a.internal.Exec(q.Raw, q.Args...)
	return res.RowsAffected, res.Error
}

func (a adapter) buildQuery(model any, q micro.Query) *gorm.DB {
	var builder *gorm.DB
	if q.Model != nil {
		builder = a.internal.Model(q.Model)
	} else {
		builder = a.internal.Model(model)
	}

	if q.Raw != "" {
		builder = builder.Raw(strings.TrimSpace(q.Raw), q.Args...)
	} else {
		if q.W != "" {
			builder = builder.Where(strings.TrimSpace(q.W), q.Args...)
		}
		if q.Sort != "" {
			builder = builder.Order(q.Sort)
		}
		if q.Select != "" {
			builder = builder.Select(q.Select)
		}
	}

	return builder
}

func (a adapter) Patch(model interface{}, id string, data map[string]interface{}) (int64, error) {
	res := a.internal.Model(model).Where("id=?", id).Updates(data)
	return res.RowsAffected, res.Error
}

func (a adapter) Ping() error {
	return a.internal.Exec("SELECT 1").Error
}

func (a adapter) NewTx() micro.DataSource {
	tx := a.internal.Begin()
	return &adapter{
		internal: tx,
		url:      a.url,
		tenantId: a.tenantId,
	}
}

func (a adapter) Transaction(cb func(tx micro.DataSource) error) error {
	return a.internal.Transaction(func(tx *gorm.DB) error {
		return cb(&adapter{
			internal: tx,
			url:      a.url,
			tenantId: a.tenantId,
		})
	})
}

func (a adapter) Close() {
	sqlDB, err := a.internal.DB()
	if err != nil {
		log.Printf("unable to close database: %s", err)
	} else {
		_ = sqlDB.Close()
	}
}

func (a adapter) Migrate(fs fs.FS, location string, migrationsTable string) {
	goose.SetBaseFS(fs)
	goose.SetTableName(migrationsTable)
	if err := goose.SetDialect(a.internal.Dialector.Name()); err != nil {
		log.Fatalf("unable to set dialect: %s", err)
	}
	cnx, err := a.internal.DB()
	dir := location
	if err = goose.Up(cnx, dir, goose.WithAllowMissing()); err != nil {
		if err.Error() == "no migration files found" {
			log.Warnf("no migration files found in %s", dir)
		} else {
			log.Fatal(err)
		}
	}
}

func NewGormAdapter(url string, schema string) micro.DataSource {
	db := createLink(url, schema)
	return &adapter{
		internal: db,
		tenantId: schema,
		url:      url,
	}
}

func createLink(url string, dbschema string) *gorm.DB {
	var dialector gorm.Dialector
	supportSchema := false
	tenantUrl := strings.ReplaceAll(url, "__tenant__", dbschema)
	if strings.HasPrefix(url, "postgres") || strings.HasPrefix(tenantUrl, "pg") || strings.HasPrefix(tenantUrl, "postgresql") {
		tenantUrl = strings.ReplaceAll(tenantUrl, "pg:", "postgres:")
		tenantUrl = strings.ReplaceAll(tenantUrl, "postgresql:", "postgres:")
		if dbschema != "" && dbschema != "public" {
			tenantUrl += "?search_path=" + dbschema
		}
		dialector = postgres.Open(tenantUrl)
		supportSchema = true
	} else if strings.HasPrefix(tenantUrl, "file:") || strings.HasSuffix(tenantUrl, ".db") {
		dialector = sqlite.Open(tenantUrl)
	} else {
		log.Fatalf("unsupported database type: %s", tenantUrl)
	}

	gdb, err := gorm.Open(dialector, &gorm.Config{
		Logger: gorm_logrus.New(),
	})

	if err == nil && supportSchema && dbschema != "" {
		err = gdb.Exec("create schema if not exists  " + dbschema).Error
	}

	if err != nil {
		log.Fatalf("unable to connect to database: %s", err)
	}

	return gdb
}
