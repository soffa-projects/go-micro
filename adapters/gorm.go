package adapters

import (
	"github.com/fabriqs/go-micro/micro"
	"github.com/fabriqs/go-micro/util/errors"
	"github.com/fabriqs/go-micro/util/h"
	_ "github.com/jackc/pgx/v5"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strings"
	"time"
)

type adapter struct {
	micro.DataSource
	tenants map[string]*gorm.DB
	current *gorm.DB
	config  *micro.DataSourceCfg
}

func (a adapter) get(ctx micro.Ctx) *gorm.DB {
	if a.current != nil {
		return a.current
	}
	tenantId := ctx.TenantId
	if tenantId == "" {
		tenantId = micro.DefaltTenantId
	}
	if cx, ok := a.tenants[tenantId]; ok {
		return cx
	} else {
		h.RaiseAny(errors.Technical("invalid_tenant: " + tenantId))
		return nil
	}
}

func (a adapter) Create(ctx micro.Ctx, model interface{}) error {
	return a.get(ctx).Create(model).Error
}

func (a adapter) Save(ctx micro.Ctx, model interface{}) error {
	return a.get(ctx).Save(model).Error
}

func (a adapter) Exists(ctx micro.Ctx, model any, q micro.Query) (bool, error) {
	var count int64
	res := a.buildQuery(ctx, model, q).Count(&count)
	return count > 0, res.Error
}

func (a adapter) Find(ctx micro.Ctx, target any, q micro.Query) error {
	res := a.buildQuery(ctx, target, q).Find(target)
	return res.Error
}

func (a adapter) First(ctx micro.Ctx, model any, q micro.Query) error {
	res := a.buildQuery(ctx, model, q).First(model)
	if res.Error == gorm.ErrRecordNotFound {
		return micro.ErrRecordNotFound
	}
	if res.RowsAffected == 0 {
		return micro.ErrRecordNotFound
	}
	return res.Error
}

func (a adapter) Count(ctx micro.Ctx, model any, q micro.Query) (int64, error) {
	var count int64
	res := a.buildQuery(ctx, model, q).Count(&count)
	return count, res.Error
}

func (a adapter) Delete(ctx micro.Ctx, model any, q micro.Query) (int64, error) {
	res := a.buildQuery(ctx, model, q).Delete(model)
	return res.RowsAffected, res.Error
}

func (a adapter) Execute(ctx micro.Ctx, model any, q micro.Query) (int64, error) {
	res := a.get(ctx).Raw(q.Raw, q.Args...).Scan(model)
	return res.RowsAffected, res.Error
}

func (a adapter) buildQuery(ctx micro.Ctx, model any, q micro.Query) *gorm.DB {
	var builder *gorm.DB
	if q.Model != nil {
		builder = a.get(ctx).Model(q.Model)
	} else {
		builder = a.get(ctx).Model(model)
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

func (a adapter) Patch(ctx micro.Ctx, model interface{}, id string, data map[string]interface{}) (int64, error) {
	res := a.get(ctx).Model(model).Where("id=?", id).Updates(data)
	return res.RowsAffected, res.Error
}

func (a adapter) Ping(ctx micro.Ctx) error {
	return a.get(ctx).Exec("SELECT 1").Error
}

func (a adapter) Transaction(ctx micro.Ctx, cb func(tx micro.DataSource) error) error {
	return a.get(ctx).Transaction(func(tx *gorm.DB) error {
		return cb(&adapter{current: tx})
	})
}

func (a adapter) Close() {
	for _, db := range a.tenants {
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("unable to close database: %s", err)
		} else {
			sqlDB.Close()
		}
	}
}

func (a adapter) Migrate() {
	a.MigrateTenant(micro.DefaltTenantId)
	if a.tenants != nil {
		for tenant := range a.tenants {
			if tenant != micro.DefaltTenantId {
				a.MigrateTenant(tenant)
			}
		}
	}
}

func (a adapter) MigrateTenants(tenant []string) {
	for _, t := range tenant {
		a.MigrateTenant(t)
	}
}

func (a adapter) MigrateTenant(tenant string) {

	db := a.get(micro.Ctx{TenantId: tenant})

	goose.SetBaseFS(a.config.Migrations)
	goose.SetTableName("z_migrations")
	goose.SetDialect(db.Dialector.Name())

	location := "shared"

	if tenant != micro.DefaltTenantId {
		location = "tenant"
	}
	cnx, err := db.DB()
	dir := location
	if err = goose.Up(cnx, dir, goose.WithAllowMissing()); err != nil {
		log.Fatal(err)
	}
}

func NewGormAdapter(config *micro.DataSourceCfg) micro.DataSource {
	start := time.Now()
	defer func() {
		log.Printf("Gorm adapter initialized in %s", time.Since(start))
	}()

	links := map[string]*gorm.DB{}
	db := createLink(config.Url, "public")

	links[micro.DefaltTenantId] = db

	if config.TenantLoader != nil {
		for _, tenant := range config.TenantLoader.GetTenant() {
			links[tenant] = createLink(config.Url, tenant)
		}
	}

	adater := &adapter{tenants: links, config: config}

	adater.Migrate()
	return adater
}

func createLink(url string, dbschema string) *gorm.DB {
	var dialector gorm.Dialector
	supportSchema := false
	if strings.HasPrefix(url, "postgres") || strings.HasPrefix(url, "pg") || strings.HasPrefix(url, "postgresql") {
		url = strings.ReplaceAll(url, "pg:", "postgres:")
		url = strings.ReplaceAll(url, "postgresql:", "postgres:")
		if dbschema != "" && dbschema != "public" {
			url += "?search_path=" + dbschema
		}
		dialector = postgres.Open(url)
		supportSchema = true
	} else if strings.HasPrefix(url, "file:") || strings.HasSuffix(url, ".db") {
		dialector = sqlite.Open(url)
	} else {
		log.Fatalf("unsupported database type: %s", url)
	}

	dbLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second * 1, // Slow SQL threshold
			LogLevel:                  logger.Silent,   // Log level
			IgnoreRecordNotFoundError: true,            // Ignore ErrRecordNotFound error for logger
			//ParameterizedQueries:      true,            // Don't include params in the SQL log
			Colorful: false, // Disable color
		},
	)

	gdb, err := gorm.Open(dialector, &gorm.Config{
		Logger: dbLogger,
	})

	if err == nil && supportSchema {
		err = gdb.Exec("create schema if not exists  " + dbschema).Error
	}

	if err != nil {
		log.Fatalf("unable to connect to database: %s", err)
	}

	return gdb
}
