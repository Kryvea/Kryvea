package db

import (
	"context"
	"errors"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type TargetIndex struct{ driver *Driver }

func (ti *TargetIndex) targetWithCustomerSelect(ctx context.Context, row any) *bun.SelectQuery {
	return idbFrom(ctx, ti.driver.db).NewSelect().Model(row).Relation("Customer")
}

func (ti *TargetIndex) Insert(ctx context.Context, target *model.Target, customerID uuid.UUID) (uuid.UUID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}
	row := &dbTarget{
		ID:       id,
		IPv4:     target.IPv4,
		IPv6:     target.IPv6,
		FQDN:     target.FQDN,
		Tag:      target.Tag,
		Protocol: target.Protocol,
		Port:     target.Port,
	}
	if customerID != uuid.Nil {
		c := customerID
		row.CustomerID = &c
	}
	if _, err := idbFrom(ctx, ti.driver.db).NewInsert().Model(row).Exec(ctx); err != nil {
		return uuid.Nil, mapErr(err)
	}
	target.ID = id
	target.Customer.ID = customerID
	return id, nil
}

func (ti *TargetIndex) FirstOrInsert(ctx context.Context, target *model.Target, customerID uuid.UUID) (uuid.UUID, bool, error) {
	var existing dbTarget
	q := idbFrom(ctx, ti.driver.db).NewSelect().
		Model(&existing).
		Column("id").
		Where("ipv4 = ?", target.IPv4).
		Where("ipv6 = ?", target.IPv6).
		Where("fqdn = ?", target.FQDN).
		Where("tag = ?", target.Tag)
	if customerID == uuid.Nil {
		q = q.Where("customer_id IS NULL")
	} else {
		q = q.Where("customer_id = ?", customerID)
	}
	if err := q.Scan(ctx); err == nil {
		return existing.ID, false, nil
	} else if !errors.Is(mapErr(err), store.ErrNotFound) {
		return uuid.Nil, false, mapErr(err)
	}
	id, err := ti.Insert(ctx, target, customerID)
	return id, true, err
}

func (ti *TargetIndex) Update(ctx context.Context, id uuid.UUID, target *model.Target) error {
	if id == model.ImmutableID {
		return store.ErrImmutableTarget
	}
	_, err := idbFrom(ctx, ti.driver.db).NewUpdate().
		Model((*dbTarget)(nil)).
		Set("ipv4 = ?", target.IPv4).
		Set("ipv6 = ?", target.IPv6).
		Set("port = ?", target.Port).
		Set("protocol = ?", target.Protocol).
		Set("fqdn = ?", target.FQDN).
		Set("tag = ?", target.Tag).
		Where("id = ?", id).
		Exec(ctx)
	return mapErr(err)
}

func (ti *TargetIndex) Delete(ctx context.Context, id uuid.UUID) error {
	if id == model.ImmutableID {
		return store.ErrImmutableTarget
	}
	_, err := idbFrom(ctx, ti.driver.db).NewDelete().
		Model((*dbTarget)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return mapErr(err)
}

func (ti *TargetIndex) GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*model.Target, error) {
	var row dbTarget
	if err := ti.targetWithCustomerSelect(ctx, &row).
		Where("t.id = ?", id).
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := row.toModel()
	return &out, nil
}

func (ti *TargetIndex) Search(ctx context.Context, customerID uuid.UUID, query string) ([]model.Target, error) {
	var rows []dbTarget
	q := ti.targetWithCustomerSelect(ctx, &rows).Where("t.id <> ?", model.ImmutableID)
	if customerID != uuid.Nil {
		q = q.Where("t.customer_id = ?", customerID)
	}
	if query != "" {
		lk := "%" + query + "%"
		q = q.Where(`(t.ipv4 ILIKE ? OR t.ipv6 ILIKE ? OR t.fqdn ILIKE ? OR t.tag ILIKE ? OR t.protocol ILIKE ? OR t.port::text ILIKE ?)`,
			lk, lk, lk, lk, lk, lk)
	}
	q = q.OrderExpr("t.fqdn, t.ipv4")
	if err := q.Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := make([]model.Target, len(rows))
	for i := range rows {
		out[i] = rows[i].toModel()
	}
	return out, nil
}
