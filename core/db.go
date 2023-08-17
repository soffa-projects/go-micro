package micro

import (
	"embed"
	serrors "errors"
	"github.com/fabriqs/go-micro/util/errors"
	"github.com/fabriqs/go-micro/util/h"
	"github.com/thoas/go-funk"
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

	Exists(model any, c Criteria) (bool, error)
	First(model any, q Q, target any) Result
	Find(model any, q Q, target any) Result

	FindBy(target any, where string, args ...any) error
	FirstBy(target any, where string, args ...any) error
	FirstBySorted(target any, orderBy string, where string, args ...any) error
	ExistBy(model any, where string, args ...any) (bool, error)
	DeleteBy(model any, where string, args ...any) (int64, error)
	Raw(target any, query string, args ...any) error
	Count(model any) (int64, error)
	CountBy(model any, where string, args ...any) (int64, error)
	Patch(model interface{}, where string, args []interface{}, data map[string]interface{}) (int64, error)
	FindAll(target any) error
	FindAllSorted(target any, orderBy string) error
	FindBySorted(target any, orderBy string, where string, args ...any) error
	DeleteAll(model any) (int64, error)
	Ping() error
}

var ErrRecordNotFound = errors.Functional("record not found")

type Criteria struct {
	W       string
	OrderBy string
	Args    []any
}

type Q struct {
	Criteria
	OrderBy string
	Select  string
	Offset  int64
	Limit   int64
}

type Result struct {
	RowsAffected int64
	Error        error
}

type Basebase[T any] interface {
	PreCreate(entity *T)
}

type Repo[T any] struct {
	Ctx           Ctx
	PreCreate     func(entity *T)
	ConflictQuery func(entity T) Criteria
	Merge         func(entity *T, existing T)
}

type Lifecyle[T any] struct {
	FindExisting func() (*T, error)
	Patch        func(existing *T, model T)
	PreCreate    func()
}

type Entity[T any] interface {
	PreCreate()
}

func (r *Criteria) Add(where string, arg ...any) {
	if r.W != "" {
		r.W += " AND "
	}
	r.W += where
	r.Args = append(r.Args, arg...)
}

func W(where string, args ...any) Criteria {
	return Criteria{W: where, Args: args}
}

func (r *Repo[T]) Count() (int64, error) {
	var model T
	return r.Ctx.DB().Count(model)
}

func (r *Repo[T]) Patch(id string, data h.Map) (int64, error) {
	var model T
	return r.Ctx.DB().Patch(model, "id=?", []any{id}, data)
}

func (r *Repo[T]) CountBy(where string, args ...any) (int64, error) {
	var model T
	return r.Ctx.DB().CountBy(model, where, args...)
}

func (r *Repo[T]) FindById(id string) (*T, error) {
	var model T
	err := r.Ctx.DB().FirstBy(&model, "id=?", id)
	if serrors.Is(err, ErrRecordNotFound) {
		return nil, nil
	}
	return &model, err
}

func (r *Repo[T]) Import(items []T, getId func(item T) string) (int, error) {
	existsing, err := r.FindAll() //TODO: fetch only ids
	if err != nil {
		return 0, err
	}

	existingIds := funk.Map(existsing, func(item *T) string {
		value := getId(*item)
		return value
	}).([]string)

	toBeAdded := funk.Filter(items, func(item T) bool {
		itemId := getId(item)
		return !funk.Contains(existingIds, itemId)
	}).([]T)

	if len(toBeAdded) > 0 {
		err = r.CreateAllImmutable(toBeAdded)
		if err != nil {
			return 0, err
		}
	}
	return len(toBeAdded), nil

}

func (r *Repo[T]) Save(id *string, model *T) error {

	if r.ConflictQuery != nil {
		conflict := r.ConflictQuery(*model)
		if exists, err := r.Exists(conflict); err != nil {
			return err
		} else if exists {
			return errors.Conflict("duplicate")
		}
	}

	if id == nil {
		return r.Create(model)
	}

	if existing, err := r.FindById(*id); err != nil {
		return err
	} else if existing == nil {
		return errors.ResourceNotFound("record_not_found")
	} else if r.Merge != nil {
		r.Merge(model, *existing)
		return r.Update(model)
	} else {
		return r.Update(model)
	}
}

func (r *Repo[T]) Update(data *T) error {
	return r.Ctx.DB().Save(&data)
}

func (r *Repo[T]) UpdateNext(data *T) error {
	err := r.Ctx.DB().Save(data)
	return err
}

func (r *Repo[T]) MergeAndSave(data *T, existing T) error {
	if r.Merge != nil {
		r.Merge(data, existing)
	}
	err := r.Ctx.DB().Save(data)
	return err
}

func (r *Repo[T]) SaveAll(data []*T) error {

	return r.Ctx.DB().Save(data)
}

func (r *Repo[T]) FetchInto(cr Criteria, target interface{}) error {
	var model T
	q := Q{Criteria: cr}
	res := r.Ctx.DB().Find(model, q, &target)
	return res.Error
}

func (r *Repo[T]) Fetch(cr Criteria) ([]T, error) {
	var target []T
	var model T
	q := Q{Criteria: cr}
	res := r.Ctx.DB().Find(model, q, &target)
	return target, res.Error
}

func (r *Repo[T]) Create(record *T) error {
	/*if e, ok := reflect.ValueOf(data).Interface().(Entity[T]); ok {
		e.PreCreate()
	}*/
	if r.PreCreate != nil {
		r.PreCreate(record)
	}
	return r.Ctx.DB().Create(&record)
}

func (r *Repo[T]) FindAll() ([]*T, error) {
	var data []*T
	err := r.Ctx.DB().FindAll(&data)
	return data, err
}

func (r *Repo[T]) CreateAll(data []*T) error {
	if r.PreCreate != nil {
		for _, d := range data {
			r.PreCreate(d)
		}
	}
	err := r.Ctx.DB().Create(data)
	return err
}

func (r *Repo[T]) CreateAllImmutable(data []T) error {
	items := funk.Map(data, func(item T) *T {
		return &item
	}).([]*T)
	if r.PreCreate != nil {
		for _, d := range items {
			r.PreCreate(d)
		}
	}
	err := r.Ctx.DB().Create(items)
	return err
}

func (r *Repo[T]) Raw(target interface{}, query string, args ...any) error {
	return r.Ctx.DB().Raw(target, query, args...)
}

func (r *Repo[T]) Exec(target interface{}, query string, args ...any) error {
	return r.Ctx.DB().Raw(target, query, args...)
}

func (r *Repo[T]) FindBySorted(orderBy string, where string, args ...any) ([]*T, error) {
	var model []*T
	err := r.Ctx.DB().FindBySorted(&model, orderBy, where, args...)
	return model, err
}

func (r *Repo[T]) FindBy(where string, args ...any) ([]*T, error) {
	var model []*T
	err := r.Ctx.DB().FindBy(&model, where, args...)
	return model, err
}

/*
func (r *Repo[T]) Merge(model *T, lc Lifecyle[T]) (*T, error) {
	if existing, err := lc.FindExisting(); err != nil {
		return nil, err
	} else if existing != nil {
		lc.Patch(existing, *model)
		err = r.Update(existing)
		return existing, err
	} else {
		lc.PreCreate()
		err = r.Create(model)
		return model, err
	}
}*/

func (r *Repo[T]) FirstBy(where string, args ...any) (*T, error) {
	model := new(T)
	if err := r.Ctx.DB().FirstBy(model, where, args...); err == ErrRecordNotFound {
		return nil, nil
	} else {
		return model, err
	}
}

func (r *Repo[T]) FirstBySorted(orderBy string, where string, args ...any) (*T, error) {
	model := new(T)
	if err := r.Ctx.DB().FirstBySorted(model, orderBy, where, args...); err == ErrRecordNotFound {
		return nil, nil
	} else {
		return model, err
	}
}

func (r *Repo[T]) FirstById(value any) (T, error) {
	var model T
	if err := r.Ctx.DB().FirstBy(&model, "id=?", value); err == ErrRecordNotFound {
		return model, nil
	} else {
		return model, err
	}
}

func (r *Repo[T]) FindAllSorted(sorted string) ([]*T, error) {
	var data []*T
	err := r.Ctx.DB().FindAllSorted(&data, sorted)
	return data, err
}

func (r *Repo[T]) DeleteById(id string) error {
	var model T
	_, err := r.Ctx.DB().DeleteBy(model, "id = ?", id)
	return err
}

func (r *Repo[T]) DeleteBy(where string, args ...interface{}) error {
	var model T
	_, err := r.Ctx.DB().DeleteBy(model, where, args...)
	return err
}

func (r *Repo[T]) Exists(cr Criteria) (bool, error) {
	var model T
	return r.Ctx.DB().Exists(model, cr)
}

func (r *Repo[T]) ExistsBy(where string, args ...interface{}) (bool, error) {
	var model T
	return r.Ctx.DB().Exists(model, W(where, args...))
}
