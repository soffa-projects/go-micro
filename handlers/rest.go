package handlers

import (
	"github.com/oleiade/reflections"
	"github.com/soffa-projects/go-micro/micro"
	"github.com/soffa-projects/go-micro/schema"
	"github.com/soffa-projects/go-micro/util/errors"
	"github.com/soffa-projects/go-micro/util/h"
	"github.com/soffa-projects/go-micro/util/ids"
)

func CRUD[Dto any, CreateDto any, UpdateDto any](g micro.BaseRouter) {
	g.GET("", func(ctx micro.Ctx, input schema.PagingInput) schema.EntityList[Dto] {
		return GetEntityList[Dto](ctx, input)
	})
	g.POST("/search", func(ctx micro.Ctx, filter schema.FilterInput) schema.EntityList[Dto] {
		return SearchEntity[Dto](ctx, filter)
	})
	g.POST("", func(ctx micro.Ctx, input CreateDto) Dto {
		var model Dto
		return CreateEntity(ctx, input, model)
	})
	g.DELETE("/:id", func(ctx micro.Ctx, input schema.IdModel) schema.IdModel {
		return DeleteEntity[Dto](ctx, input)
	})
	g.PATCH("/:id", func(ctx micro.Ctx, input UpdateDto) Dto {
		return UpdateEntity[Dto](ctx, input)
	})
}

func GetEntityList[T any](c micro.Ctx, paging schema.PagingInput) schema.EntityList[T] {
	db := c.CurrentDB()
	var data []T
	page := 1
	limit := 1000
	if paging.Count > 0 {
		limit = paging.Count
	}
	if paging.Page > 1 {
		page = paging.Page
	}
	q := micro.Query{
		Offset: int64((page - 1) * limit),
		Limit:  int64(limit),
	}
	h.RaiseAny(db.Find(&data, q))
	return schema.EntityList[T]{
		Data: data,
	}
}

func SearchEntity[T any](c micro.Ctx, input schema.FilterInput) schema.EntityList[T] {
	db := c.CurrentDB()
	var data []T
	page := 1
	limit := 1000
	if input.Count > 0 {
		limit = input.Count
	}
	if input.Page > 1 {
		page = input.Page
	}
	q := micro.Query{
		W:      input.Where,
		Args:   input.Args,
		Offset: int64((page - 1) * limit),
		Limit:  int64(limit),
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
