package handlers

import (
	"github.com/oleiade/reflections"
	log "github.com/sirupsen/logrus"
	"github.com/soffa-projects/go-micro/micro"
	"github.com/soffa-projects/go-micro/schema"
	"github.com/soffa-projects/go-micro/util/errors"
	"github.com/soffa-projects/go-micro/util/h"
	"github.com/soffa-projects/go-micro/util/ids"
)

func GetEntityList[T any](c micro.Ctx, filter ...schema.FilterInput) schema.EntityList[T] {
	db := c.CurrentDB()
	var data []T
	q := micro.Query{}
	if len(filter) > 0 {
		q.W = filter[0].Where
		q.Args = filter[0].Args
	}
	h.RaiseAny(db.Find(&data, q))
	return schema.EntityList[T]{
		Data: data,
	}
}

func FilterEntityList[T any](c micro.Ctx, criteria any) schema.EntityList[T] {
	db := c.CurrentDB()
	var data []T
	q := micro.Query{}
	if criteria != nil {
		fields := h.F(h.ToMap(criteria))
		if len(fields) > 0 {
			q.W = "1=1"
			for k, v := range fields {
				if !h.IsEmpty(v) {
					q.W += " AND " + k + " = ?"
					q.Args = append(q.Args, v)
				}
			}
		}
	}
	url := c.Request().URL.Query()
	page := url.Get("page")
	limit := url.Get("limit")
	if page != "" && limit != "" {
		log.Debugf("fetchin paginated data: page=%v, limit=%v", page, limit)
	}
	h.RaiseAny(db.Find(&data, q))
	return schema.EntityList[T]{
		Data: data,
	}
}

func CreateEntity[T any](c micro.Ctx, input any, entity T) T {
	db := c.CurrentDB()
	//var entity T
	h.RaiseAny(h.CopyAllFields(&entity, input, true))
	prefix := h.F(reflections.GetFieldTag(entity, "Id", "prefix"))
	h.RaiseIf(h.IsStrEmpty(prefix), errors.Technical("entity_missing_id_prefix"))
	h.RaiseAny(reflections.SetField(&entity, "Id", ids.NewIdPtr(prefix)))
	h.RaiseAny(db.Create(&entity))
	return entity
}

func UpdateEntity[T any](c micro.Ctx, input any) T {
	db := c.CurrentDB()
	id := h.UnwrapStr(h.F(reflections.GetField(input, "Id")))
	var entity T
	err := db.First(&entity, micro.Query{
		W:    "id = ?",
		Args: []any{id},
	})
	h.RaiseAny(err)
	h.RaiseAny(h.CopyAllFields(&entity, input, true))
	h.RaiseAny(db.Save(&entity))
	return entity
}

func DeleteEntity[T any](c micro.Ctx, input schema.IdModel) schema.IdModel {
	db := c.CurrentDB()
	var entity T
	_, err := db.Delete(entity, micro.Query{
		W:    "id = ?",
		Args: []any{*input.Id},
	})
	h.RaiseAny(err)
	return input
}
