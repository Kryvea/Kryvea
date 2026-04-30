package db

import (
	"context"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type CustomerIndex struct{ driver *Driver }

func (ci *CustomerIndex) Insert(ctx context.Context, customer *model.Customer) (uuid.UUID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	if customer.LogoID != uuid.Nil {
		customer.LogoReference = util.CreateImageReference(customer.LogoMimeType, customer.LogoID)
	}

	row := &dbCustomer{
		ID:            id,
		Name:          customer.Name,
		Language:      customer.Language,
		LogoMimeType:  customer.LogoMimeType,
		LogoReference: customer.LogoReference,
	}
	if customer.LogoID != uuid.Nil {
		l := customer.LogoID
		row.LogoID = &l
	}

	if _, err := idbFrom(ctx, ci.driver.db).NewInsert().Model(row).Exec(ctx); err != nil {
		return uuid.Nil, mapErr(err)
	}

	customer.ID = id
	return id, nil
}

func (ci *CustomerIndex) Update(ctx context.Context, id uuid.UUID, customer *model.Customer) error {
	_, err := idbFrom(ctx, ci.driver.db).NewUpdate().
		Model((*dbCustomer)(nil)).
		Set("name = ?", customer.Name).
		Set("language = ?", customer.Language).
		Where("id = ?", id).
		Exec(ctx)
	return mapErr(err)
}

func (ci *CustomerIndex) UpdateLogo(ctx context.Context, id, logoID uuid.UUID, mime string) error {
	q := idbFrom(ctx, ci.driver.db).NewUpdate().
		Model((*dbCustomer)(nil)).
		Set("logo_mime_type = ?", mime).
		Set("logo_reference = ?", util.CreateImageReference(mime, logoID)).
		Where("id = ?", id)
	if logoID == uuid.Nil {
		q = q.Set("logo_id = NULL")
	} else {
		q = q.Set("logo_id = ?", logoID)
	}
	_, err := q.Exec(ctx)
	return mapErr(err)
}

func (ci *CustomerIndex) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := idbFrom(ctx, ci.driver.db).NewDelete().
		Model((*dbCustomer)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return mapErr(err)
}

func (ci *CustomerIndex) GetByID(ctx context.Context, id uuid.UUID) (*model.Customer, error) {
	var row dbCustomer
	if err := idbFrom(ctx, ci.driver.db).NewSelect().
		Model(&row).
		Where("id = ?", id).
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := row.toModel()
	return &out, nil
}

func (ci *CustomerIndex) GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*model.Customer, error) {
	var row dbCustomer
	if err := idbFrom(ctx, ci.driver.db).NewSelect().
		Model(&row).
		Relation("Templates", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("created_at DESC")
		}).
		Where("c.id = ?", id).
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := row.toModel()
	out.Templates = make([]model.Template, len(row.Templates))
	for i := range row.Templates {
		out.Templates[i] = row.Templates[i].toModelBare()
	}
	return &out, nil
}

func (ci *CustomerIndex) GetAll(ctx context.Context, ids []uuid.UUID) ([]model.Customer, error) {
	var rows []dbCustomer
	q := idbFrom(ctx, ci.driver.db).NewSelect().
		Model(&rows).
		Relation("Templates", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("created_at DESC")
		}).
		Order("c.name")
	if ids != nil {
		q = q.Where("c.id IN (?)", bun.List(ids))
	}
	if err := q.Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := make([]model.Customer, len(rows))
	for i := range rows {
		out[i] = rows[i].toModel()
		out[i].Templates = make([]model.Template, len(rows[i].Templates))
		for j := range rows[i].Templates {
			out[i].Templates[j] = rows[i].Templates[j].toModelBare()
		}
	}
	return out, nil
}
