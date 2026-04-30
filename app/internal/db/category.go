package db

import (
	"context"
	"errors"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/google/uuid"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type CategoryIndex struct{ driver *Driver }

func (ci *CategoryIndex) Insert(ctx context.Context, category *model.Category) (uuid.UUID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}
	row := &dbCategory{
		ID:                 id,
		Identifier:         category.Identifier,
		Name:               category.Name,
		Subcategory:        category.Subcategory,
		GenericDescription: emptyMapIfNil(category.GenericDescription),
		GenericRemediation: emptyMapIfNil(category.GenericRemediation),
		LanguagesOrder:     emptyStringsIfNil(category.LanguagesOrder),
		References:         emptyStringsIfNil(category.References),
		Source:             category.Source,
	}
	if _, err := idbFrom(ctx, ci.driver.db).NewInsert().Model(row).Exec(ctx); err != nil {
		return uuid.Nil, mapErr(err)
	}
	category.ID = id
	return id, nil
}

func (ci *CategoryIndex) Upsert(ctx context.Context, category *model.Category, override bool) (uuid.UUID, error) {
	if !override {
		return ci.Insert(ctx, category)
	}
	id, isNew, err := ci.FirstOrInsert(ctx, category)
	if err != nil {
		return uuid.Nil, err
	}
	if isNew {
		return id, nil
	}
	if err := ci.Update(ctx, id, category); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (ci *CategoryIndex) FirstOrInsert(ctx context.Context, category *model.Category) (uuid.UUID, bool, error) {
	var row dbCategory
	err := idbFrom(ctx, ci.driver.db).NewSelect().
		Model(&row).
		Column("id").
		Where("identifier = ?", category.Identifier).
		Where("name = ?", category.Name).
		Where("subcategory = ?", category.Subcategory).
		Scan(ctx)
	if err == nil {
		return row.ID, false, nil
	}
	if !errors.Is(mapErr(err), store.ErrNotFound) {
		return uuid.Nil, false, mapErr(err)
	}
	id, err := ci.Insert(ctx, category)
	return id, true, err
}

func (ci *CategoryIndex) Update(ctx context.Context, id uuid.UUID, category *model.Category) error {
	if id == model.ImmutableID {
		return store.ErrImmutableCategory
	}
	_, err := idbFrom(ctx, ci.driver.db).NewUpdate().
		Model((*dbCategory)(nil)).
		Set("identifier = ?", category.Identifier).
		Set("name = ?", category.Name).
		Set("subcategory = ?", category.Subcategory).
		Set("generic_description = ?", emptyMapIfNil(category.GenericDescription)).
		Set("generic_remediation = ?", emptyMapIfNil(category.GenericRemediation)).
		Set("languages_order = ?", pgdialect.Array(emptyStringsIfNil(category.LanguagesOrder))).
		Set("refs = ?", pgdialect.Array(emptyStringsIfNil(category.References))).
		Set("source = ?", category.Source).
		Where("id = ?", id).
		Exec(ctx)
	return mapErr(err)
}

func (ci *CategoryIndex) Delete(ctx context.Context, id uuid.UUID) error {
	if id == model.ImmutableID {
		return store.ErrImmutableCategory
	}
	_, err := idbFrom(ctx, ci.driver.db).NewDelete().
		Model((*dbCategory)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return mapErr(err)
}

func (ci *CategoryIndex) GetAll(ctx context.Context) ([]model.Category, error) {
	var rows []dbCategory
	if err := idbFrom(ctx, ci.driver.db).NewSelect().
		Model(&rows).
		Where("id <> ?", model.ImmutableID).
		Order("name").
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := make([]model.Category, len(rows))
	for i, r := range rows {
		out[i] = r.toModel()
	}
	return out, nil
}

func (ci *CategoryIndex) GetByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	var row dbCategory
	if err := idbFrom(ctx, ci.driver.db).NewSelect().
		Model(&row).
		Where("id = ?", id).
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := row.toModel()
	return &out, nil
}

func (ci *CategoryIndex) Search(ctx context.Context, query string) ([]model.Category, error) {
	if query == "" {
		return ci.GetAll(ctx)
	}
	like := "%" + query + "%"
	var rows []dbCategory
	if err := idbFrom(ctx, ci.driver.db).NewSelect().
		Model(&rows).
		Where("id <> ?", model.ImmutableID).
		Where("(identifier ILIKE ? OR name ILIKE ? OR subcategory ILIKE ?)", like, like, like).
		Order("name").
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := make([]model.Category, len(rows))
	for i, r := range rows {
		out[i] = r.toModel()
	}
	return out, nil
}

func emptyMapIfNil(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return m
}

func emptyStringsIfNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
