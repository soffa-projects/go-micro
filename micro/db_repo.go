package micro

import (
	serrors "errors"
	"github.com/soffa-projects/go-micro/util/errors"
	"github.com/soffa-projects/go-micro/util/h"
	"github.com/thoas/go-funk"
)

type RepoHooks[T any] struct {
	PreCreate func(e *T)
}

type EntityRepo[T any] interface {
	Create(Ctx, *T) error
	CreateAll(Ctx, []*T) error
	UpdateAll(Ctx, []*T) error
	Update(Ctx, *T) error
	DeleteById(Ctx, string) error
	Merge(Ctx, string, func(target *T)) (*T, error)
	Import(Ctx, []*T, func(item *T) string) (int, error)
	FindAll(Ctx) ([]*T, error)
	FindAllSorted(Ctx, string) ([]*T, error)
	FindById(Ctx, string) (*T, error)
	FindByIds(ctx Ctx, values []string) ([]*T, error)
	CountAll(Ctx) (int64, error)
}

type EntityRepoImpl[T any] interface {
	EntityRepo[T]
	Query(Ctx, interface{}, string, ...interface{}) error
	Raw(Ctx, string, ...interface{}) error
	Patch(Ctx, string, map[string]any) error
	FindBySorted(ctx Ctx, sortBy string, where string, args ...interface{}) ([]*T, error)
	FindBy(Ctx, string, ...interface{}) ([]*T, error)
	FindByInto(Ctx, any, string, ...interface{}) error
	FirstBy(Ctx, string, ...interface{}) (*T, error)
	CountBy(Ctx, string, ...interface{}) (int64, error)
	ExistsBy(Ctx, string, ...interface{}) (bool, error)
	DeleteBy(Ctx, string, ...interface{}) error
}

type entityRepoImpl[T any] struct {
	EntityRepo[T]
	//db    DataSource
	hooks RepoHooks[T]
}

func NewRepoImpl[T any](preCreate func(e *T)) EntityRepoImpl[T] {
	return entityRepoImpl[T]{
		hooks: RepoHooks[T]{
			PreCreate: preCreate,
		},
	}
}

// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// COMMANDS
// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

func (r entityRepoImpl[T]) CreateAll(ctx Ctx, entities []*T) error {
	for _, e := range entities {
		r.hooks.PreCreate(e)
	}
	return ctx.db.Create(entities)
}

func (r entityRepoImpl[T]) Create(ctx Ctx, record *T) error {
	r.hooks.PreCreate(record)
	return ctx.db.Create(&record)
}

func (r entityRepoImpl[T]) Update(ctx Ctx, data *T) error {
	return ctx.db.Save(&data)
}

func (r entityRepoImpl[T]) UpdateAll(ctx Ctx, data []*T) error {
	return ctx.db.Save(&data)
}

func (r entityRepoImpl[T]) DeleteBy(ctx Ctx, where string, args ...interface{}) error {
	var model T
	_, err := ctx.db.Delete(&model, Query{W: where, Args: args})
	return err
}

func (r entityRepoImpl[T]) DeleteById(ctx Ctx, value string) error {
	return r.DeleteBy(ctx, "id=?", value)
}

func (r entityRepoImpl[T]) Patch(ctx Ctx, id string, value map[string]interface{}) error {
	var model T
	_, err := ctx.db.Patch(model, id, value)
	return err
}

func (r entityRepoImpl[T]) Merge(ctx Ctx, id string, merger func(target *T)) (*T, error) {
	loaded, err := r.FindById(ctx, id)
	if err != nil {
		return nil, errors.ResourceNotFound("missing_entity")
	}
	beforeMerge := *loaded
	merger(loaded)
	if &beforeMerge != loaded {
		err = ctx.db.Save(loaded)
	}
	return loaded, err
}

func (r entityRepoImpl[T]) Import(ctx Ctx, items []*T, getId func(item *T) string) (int, error) {

	existsing, err := r.FindAll(ctx) //TODO: fetch only ids

	if err != nil {
		return 0, err
	}

	existingIds := funk.Map(existsing, func(item *T) string {
		value := getId(item)
		return value
	}).([]string)

	toBeAdded := funk.Filter(items, func(item *T) bool {
		itemId := getId(item)
		return !funk.Contains(existingIds, itemId)
	}).([]*T)

	if len(toBeAdded) > 0 {
		for _, item := range toBeAdded {
			r.hooks.PreCreate(item)
		}
		err = r.CreateAll(ctx, toBeAdded)
		if err != nil {
			return 0, err
		}
	}
	return len(toBeAdded), nil
}

// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// QUERIES
// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

func (r entityRepoImpl[T]) ExistsBy(ctx Ctx, where string, args ...interface{}) (bool, error) {
	var model T
	return ctx.db.Exists(model, Query{W: where, Args: args})
}

func (r entityRepoImpl[T]) FindAll(ctx Ctx) ([]*T, error) {
	var model []*T
	err := ctx.db.Find(&model, Query{})
	return model, err
}

func (r entityRepoImpl[T]) FindAllSorted(ctx Ctx, orderBy string) ([]*T, error) {
	var model []*T
	err := ctx.db.Find(&model, Query{Sort: orderBy})
	return model, err
}

func (r entityRepoImpl[T]) FindByInto(ctx Ctx, target any, where string, args ...interface{}) error {
	var model []*T
	err := ctx.db.Find(&target, Query{W: where, Args: args, Model: model})
	return err
}

func (r entityRepoImpl[T]) FindBy(ctx Ctx, where string, args ...interface{}) ([]*T, error) {
	var model []*T
	err := ctx.db.Find(&model, Query{W: where, Args: args})
	return model, err
}

func (r entityRepoImpl[T]) FindBySorted(ctx Ctx, sort string, where string, args ...interface{}) ([]*T, error) {
	var model []*T
	err := ctx.db.Find(&model, Query{W: where, Args: args, Sort: sort})
	return model, err
}

func (r entityRepoImpl[T]) FindById(ctx Ctx, id string) (*T, error) {
	return r.FirstBy(ctx, "id=?", id)
}

func (r entityRepoImpl[T]) FindByIds(ctx Ctx, ids []string) ([]*T, error) {
	return r.FindBy(ctx, "id in (?)", ids)
}

func (r entityRepoImpl[T]) FirstBy(ctx Ctx, where string, args ...interface{}) (*T, error) {
	var model T
	err := ctx.db.First(&model, Query{W: where, Args: args})
	if serrors.Is(err, ErrRecordNotFound) {
		return nil, nil
	}
	return &model, err
}

func (r entityRepoImpl[T]) CountBy(ctx Ctx, where string, args ...interface{}) (int64, error) {
	var model T
	return ctx.db.Count(model, Query{W: where, Args: args})
}

func (r entityRepoImpl[T]) CountAll(ctx Ctx) (int64, error) {
	var model T
	return ctx.db.Count(model, Query{})
}

func (r entityRepoImpl[T]) Query(ctx Ctx, target interface{}, raw string, args ...interface{}) error {
	if !h.IsPointer(target) {
		panic("target must be a pointer")
	}
	return ctx.db.Find(target, Query{
		Raw:  raw,
		Args: args,
	})
}

func (r entityRepoImpl[T]) Raw(ctx Ctx, raw string, args ...interface{}) error {
	var m = new(T)
	_, err := ctx.db.Execute(m, Query{Model: m, Raw: raw, Args: args})
	return err
}
