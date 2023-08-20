package micro

import (
	serrors "errors"
	"github.com/fabriqs/go-micro/util/errors"
	"github.com/thoas/go-funk"
)

type RepoHooks[T any] struct {
	PreCreate func(e *T)
}

type EntityRepo[T any] interface {
	Create(*T) error
	CreateAll([]*T) error
	UpdateAll([]*T) error
	Update(*T) error
	DeleteById(string) error
	Merge(string, func(target *T)) (*T, error)
	Import([]*T, func(item *T) string) (int, error)
	FindAll() ([]*T, error)
	FindAllSorted(string) ([]*T, error)
	FindById(string) (*T, error)
	FindByIds(values []string) ([]*T, error)
	CountAll() (int64, error)
}

type EntityRepoImpl[T any] interface {
	EntityRepo[T]
	Query(interface{}, string, ...interface{}) error
	Raw(string, ...interface{}) error
	Patch(string, map[string]any) error
	FindBySorted(sortBy string, where string, args ...interface{}) ([]*T, error)
	FindBy(string, ...interface{}) ([]*T, error)
	FindByInto(any, string, ...interface{}) error
	FirstBy(string, ...interface{}) (*T, error)
	CountBy(string, ...interface{}) (int64, error)
	ExistsBy(string, ...interface{}) (bool, error)
	DeleteBy(string, ...interface{}) error
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

func (r entityRepoImpl[T]) CreateAll(entities []*T) error {
	for _, e := range entities {
		r.hooks.PreCreate(e)
	}
	return currentDB().Create(entities)
}

func (r entityRepoImpl[T]) Create(record *T) error {
	r.hooks.PreCreate(record)
	return currentDB().Create(&record)
}

func (r entityRepoImpl[T]) Update(data *T) error {
	return currentDB().Save(&data)
}

func (r entityRepoImpl[T]) UpdateAll(data []*T) error {
	return currentDB().Save(&data)
}

func (r entityRepoImpl[T]) DeleteBy(where string, args ...interface{}) error {
	var model T
	_, err := currentDB().Delete(&model, Query{W: where, Args: args})
	return err
}

func (r entityRepoImpl[T]) DeleteById(value string) error {
	return r.DeleteBy("id=?", value)
}

func (r entityRepoImpl[T]) Patch(id string, value map[string]interface{}) error {
	var model T
	_, err := currentDB().Patch(model, id, value)
	return err
}

func (r entityRepoImpl[T]) Merge(id string, merger func(target *T)) (*T, error) {
	loaded, err := r.FindById(id)
	if err != nil {
		return nil, errors.ResourceNotFound("missing_entity")
	}
	beforeMerge := *loaded
	merger(loaded)
	if &beforeMerge != loaded {
		err = currentDB().Save(loaded)
	}
	return loaded, err
}

func (r entityRepoImpl[T]) Import(items []*T, getId func(item *T) string) (int, error) {

	existsing, err := r.FindAll() //TODO: fetch only ids

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
		err = r.CreateAll(toBeAdded)
		if err != nil {
			return 0, err
		}
	}
	return len(toBeAdded), nil
}

// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// QUERIES
// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

func (r entityRepoImpl[T]) ExistsBy(where string, args ...interface{}) (bool, error) {
	var model T
	return currentDB().Exists(model, Query{W: where, Args: args})
}

func (r entityRepoImpl[T]) FindAll() ([]*T, error) {
	var model []*T
	err := currentDB().Find(&model, Query{})
	return model, err
}

func (r entityRepoImpl[T]) FindAllSorted(orderBy string) ([]*T, error) {
	var model []*T
	err := currentDB().Find(&model, Query{Sort: orderBy})
	return model, err
}

func (r entityRepoImpl[T]) FindByInto(target any, where string, args ...interface{}) error {
	var model []*T
	err := currentDB().Find(&target, Query{W: where, Args: args, Model: model})
	return err
}

func (r entityRepoImpl[T]) FindBy(where string, args ...interface{}) ([]*T, error) {
	var model []*T
	err := currentDB().Find(&model, Query{W: where, Args: args})
	return model, err
}

func (r entityRepoImpl[T]) FindBySorted(sort string, where string, args ...interface{}) ([]*T, error) {
	var model []*T
	err := currentDB().Find(&model, Query{W: where, Args: args, Sort: sort})
	return model, err
}

func (r entityRepoImpl[T]) FindById(id string) (*T, error) {
	return r.FirstBy("id=?", id)
}

func (r entityRepoImpl[T]) FindByIds(ids []string) ([]*T, error) {
	return r.FindBy("id in (?)", ids)
}

func (r entityRepoImpl[T]) FirstBy(where string, args ...interface{}) (*T, error) {
	var model T
	err := currentDB().First(&model, Query{W: where, Args: args})
	if serrors.Is(err, ErrRecordNotFound) {
		return nil, nil
	}
	return &model, err
}

func (r entityRepoImpl[T]) CountBy(where string, args ...interface{}) (int64, error) {
	var model T
	return currentDB().Count(model, Query{W: where, Args: args})
}

func (r entityRepoImpl[T]) CountAll() (int64, error) {
	var model T
	return currentDB().Count(model, Query{})
}

func (r entityRepoImpl[T]) Query(target interface{}, raw string, args ...interface{}) error {
	var m = new(T)
	return currentDB().Find(target, Query{
		Model: m,
		Raw:   raw,
		Args:  args,
	})
}

func (r entityRepoImpl[T]) Raw(raw string, args ...interface{}) error {
	var m = new(T)
	_, err := currentDB().Execute(m, Query{Model: m, Raw: raw, Args: args})
	return err
}
