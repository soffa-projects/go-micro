package db

import (
	"embed"
	serrors "errors"
	"github.com/fabriqs/go-micro/errors"
	"github.com/fabriqs/go-micro/h"
)

type Cfg struct {
	Url        string
	Migrations embed.FS
}

type DB interface {
	Transaction(func(tx DB) error) error
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
	Query(target any, query string, args ...any) error
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

type Repo[T any] struct {
	DB            DB
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

func (r *Repo[T]) Tx(cb func(tx DB) error) error {
	return r.DB.Transaction(cb)
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
	return r.DB.Count(model)
}

func (r *Repo[T]) Patch(id string, data h.Map) (int64, error) {
	var model T
	return r.DB.Patch(model, "id=?", []any{id}, data)
}

func (r *Repo[T]) CountBy(where string, args ...any) (int64, error) {
	var model T
	return r.DB.CountBy(model, where, args...)
}

func (r *Repo[T]) FindById(id string) (*T, error) {
	var model T
	err := r.DB.FirstBy(&model, "id=?", id)
	if serrors.Is(err, ErrRecordNotFound) {
		return nil, nil
	}
	return &model, err
}

func (r *Repo[T]) Save(id *string, model T) (T, error) {

	if r.ConflictQuery != nil {
		conflict := r.ConflictQuery(model)
		if exists, err := r.Exists(conflict); err != nil {
			return model, err
		} else if exists {
			return model, errors.Conflict("duplicate")
		}
	}

	if id == nil {
		return r.Create(model)
	}

	if existing, err := r.FindById(*id); err != nil {
		return model, err
	} else if existing == nil {
		return model, errors.ResourceNotFound("record_not_found")
	} else if r.Merge != nil {
		r.Merge(&model, *existing)
		return r.Update(model)
	} else {
		return r.Update(model)
	}
}

func (r *Repo[T]) Update(data T) (T, error) {
	err := r.DB.Save(&data)
	return data, err
}

func (r *Repo[T]) MergeAndSave(data T, existing T) (T, error) {
	if r.Merge != nil {
		r.Merge(&data, existing)
	}
	err := r.DB.Save(&data)
	return data, err
}

func (r *Repo[T]) SaveAll(data []T) error {
	return r.DB.Save(data)
}

func (r *Repo[T]) FetchInto(cr Criteria, target interface{}) error {
	var model T
	q := Q{Criteria: cr}
	res := r.DB.Find(model, q, &target)
	return res.Error
}

func (r *Repo[T]) Fetch(cr Criteria) ([]T, error) {
	var target []T
	var model T
	q := Q{Criteria: cr}
	res := r.DB.Find(model, q, &target)
	return target, res.Error
}

func (r *Repo[T]) Create(record T) (T, error) {
	/*if e, ok := reflect.ValueOf(data).Interface().(Entity[T]); ok {
		e.PreCreate()
	}*/
	if r.PreCreate != nil {
		r.PreCreate(&record)
	}
	err := r.DB.Create(&record)
	return record, err
}

func (r *Repo[T]) FindAll() ([]T, error) {
	var data []T
	err := r.DB.FindAll(&data)
	return data, err
}

func (r *Repo[T]) Query(target interface{}, query string, args ...any) error {
	return r.DB.Query(target, query, args...)
}

func (r *Repo[T]) FindBySorted(orderBy string, where string, args ...any) ([]T, error) {
	var model []T
	err := r.DB.FindBySorted(&model, orderBy, where, args...)
	return model, err
}

func (r *Repo[T]) FindBy(where string, args ...any) ([]T, error) {
	var model []T
	err := r.DB.FindBy(&model, where, args...)
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
	if err := r.DB.FirstBy(model, where, args...); err == ErrRecordNotFound {
		return nil, nil
	} else {
		return model, err
	}
}

func (r *Repo[T]) FirstBySorted(orderBy string, where string, args ...any) (*T, error) {
	model := new(T)
	if err := r.DB.FirstBySorted(model, orderBy, where, args...); err == ErrRecordNotFound {
		return nil, nil
	} else {
		return model, err
	}
}

func (r *Repo[T]) FirstById(value any) (T, error) {
	var model T
	if err := r.DB.FirstBy(&model, "id=?", value); err == ErrRecordNotFound {
		return model, nil
	} else {
		return model, err
	}
}

func (r *Repo[T]) FindAllSorted(sorted string) ([]T, error) {
	var data []T
	err := r.DB.FindAllSorted(&data, sorted)
	return data, err
}

func (r *Repo[T]) DeleteById(id string) error {
	var model T
	_, err := r.DB.DeleteBy(model, "id = ?", id)
	return err
}

func (r *Repo[T]) DeleteBy(where string, args ...interface{}) error {
	var model T
	_, err := r.DB.DeleteBy(model, where, args...)
	return err
}

func (r *Repo[T]) Exists(cr Criteria) (bool, error) {
	var model T
	return r.DB.Exists(model, cr)
}

func (r *Repo[T]) ExistsBy(where string, args ...interface{}) (bool, error) {
	var model T
	return r.DB.Exists(model, W(where, args...))
}
