package db

import (
	"context"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type TemplateIndex struct{ driver *Driver }

func (ti *TemplateIndex) templateWithCustomerSelect(ctx context.Context, row any) *bun.SelectQuery {
	return idbFrom(ctx, ti.driver.db).NewSelect().Model(row).Relation("Customer")
}

func (ti *TemplateIndex) Insert(ctx context.Context, template *model.Template) (uuid.UUID, error) {
	if template.FileID == uuid.Nil {
		return uuid.Nil, store.ErrTemplateFileIDRequired
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}
	row := &dbTemplate{
		ID:           id,
		Name:         template.Name,
		Filename:     template.Filename,
		Language:     template.Language,
		TemplateType: template.TemplateType,
		MimeType:     template.MimeType,
		Identifier:   template.Identifier,
		FileID:       template.FileID,
	}
	if template.Customer != nil && template.Customer.ID != uuid.Nil {
		c := template.Customer.ID
		row.CustomerID = &c
	}
	if _, err := idbFrom(ctx, ti.driver.db).NewInsert().Model(row).Exec(ctx); err != nil {
		return uuid.Nil, mapErr(err)
	}
	template.ID = id
	return id, nil
}

func (ti *TemplateIndex) GetByID(ctx context.Context, id uuid.UUID) (*model.Template, error) {
	var row dbTemplate
	if err := ti.templateWithCustomerSelect(ctx, &row).
		Where("tpl.id = ?", id).
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := row.toModel()
	return &out, nil
}

func (ti *TemplateIndex) GetByFileID(ctx context.Context, fileID uuid.UUID) (*model.Template, error) {
	var row dbTemplate
	if err := ti.templateWithCustomerSelect(ctx, &row).
		Where("tpl.file_id = ?", fileID).
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := row.toModel()
	return &out, nil
}

func (ti *TemplateIndex) GetAll(ctx context.Context) ([]model.Template, error) {
	var rows []dbTemplate
	if err := ti.templateWithCustomerSelect(ctx, &rows).
		OrderExpr("tpl.created_at DESC").
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := make([]model.Template, len(rows))
	for i := range rows {
		out[i] = rows[i].toModel()
	}
	return out, nil
}

func (ti *TemplateIndex) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := idbFrom(ctx, ti.driver.db).NewDelete().
		Model((*dbTemplate)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return mapErr(err)
}
