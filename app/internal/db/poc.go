package db

import (
	"context"
	"errors"
	"time"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PocIndex struct{ driver *Driver }

func (pi *PocIndex) Upsert(ctx context.Context, poc *model.Poc) error {
	old, err := pi.GetByVulnerabilityID(ctx, poc.VulnerabilityID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return err
	}

	for i := range poc.Pocs {
		if poc.Pocs[i].ID == uuid.Nil {
			id, err := uuid.NewRandom()
			if err != nil {
				return err
			}
			poc.Pocs[i].ID = id
		}
		if poc.Pocs[i].ImageID != uuid.Nil {
			poc.Pocs[i].ImageReference = util.CreateImageReference(poc.Pocs[i].ImageMimeType, poc.Pocs[i].ImageID)
		}
	}

	idb := idbFrom(ctx, pi.driver.db)

	var pocID uuid.UUID
	if old != nil && old.ID != uuid.Nil {
		pocID = old.ID
		if _, err := idb.NewUpdate().
			Model((*dbPoc)(nil)).
			Set("items = ?", poc.Pocs).
			Where("id = ?", pocID).
			Exec(ctx); err != nil {
			return mapErr(err)
		}
	} else {
		pocID, err = uuid.NewRandom()
		if err != nil {
			return err
		}
		if _, err := idb.NewInsert().
			Model(&dbPoc{
				ID:              pocID,
				VulnerabilityID: poc.VulnerabilityID,
				Items:           poc.Pocs,
			}).
			Exec(ctx); err != nil {
			return mapErr(err)
		}
	}
	poc.ID = pocID
	poc.UpdatedAt = time.Now()

	keptItemIDs := make([]uuid.UUID, 0, len(poc.Pocs))
	rows := make([]dbPocImage, 0, len(poc.Pocs))
	for _, item := range poc.Pocs {
		if item.ImageID == uuid.Nil {
			continue
		}
		keptItemIDs = append(keptItemIDs, item.ID)
		rows = append(rows, dbPocImage{
			PocID:           pocID,
			PocItemID:       item.ID,
			FileReferenceID: item.ImageID,
		})
	}
	if len(rows) > 0 {
		if _, err := idb.NewInsert().
			Model(&rows).
			On("CONFLICT (poc_id, poc_item_id) DO UPDATE").
			Set("file_reference_id = EXCLUDED.file_reference_id").
			Exec(ctx); err != nil {
			return mapErr(err)
		}
	}
	deleteQ := idb.NewDelete().
		Model((*dbPocImage)(nil)).
		Where("poc_id = ?", pocID)
	if len(keptItemIDs) > 0 {
		deleteQ = deleteQ.Where("poc_item_id NOT IN (?)", bun.List(keptItemIDs))
	}
	if _, err := deleteQ.Exec(ctx); err != nil {
		return mapErr(err)
	}
	return nil
}

func (pi *PocIndex) BulkInsertNew(ctx context.Context, pocs []model.Poc) error {
	if len(pocs) == 0 {
		return nil
	}
	rows := make([]dbPoc, len(pocs))
	for i, poc := range pocs {
		pocID, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		items := make([]model.PocItem, len(poc.Pocs))
		for j, item := range poc.Pocs {
			if item.ID == uuid.Nil {
				itemID, err := uuid.NewRandom()
				if err != nil {
					return err
				}
				item.ID = itemID
			}
			items[j] = item
		}
		rows[i] = dbPoc{
			ID:              pocID,
			VulnerabilityID: poc.VulnerabilityID,
			Items:           items,
		}
	}
	_, err := idbFrom(ctx, pi.driver.db).NewInsert().Model(&rows).Exec(ctx)
	return mapErr(err)
}

func (pi *PocIndex) GetByID(ctx context.Context, id uuid.UUID) (*model.Poc, error) {
	var row dbPoc
	err := idbFrom(ctx, pi.driver.db).NewSelect().
		Model(&row).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(mapErr(err), store.ErrNotFound) {
			return &model.Poc{Pocs: []model.PocItem{}}, nil
		}
		return nil, mapErr(err)
	}
	out := row.toModel()
	return &out, nil
}

func (pi *PocIndex) GetByVulnerabilityID(ctx context.Context, vulnerabilityID uuid.UUID) (*model.Poc, error) {
	var row dbPoc
	err := idbFrom(ctx, pi.driver.db).NewSelect().
		Model(&row).
		Where("vulnerability_id = ?", vulnerabilityID).
		Scan(ctx)
	if err != nil {
		if errors.Is(mapErr(err), store.ErrNotFound) {
			return &model.Poc{Pocs: []model.PocItem{}}, nil
		}
		return nil, mapErr(err)
	}
	out := row.toModel()
	return &out, nil
}

func (pi *PocIndex) GetByImageID(ctx context.Context, imageID uuid.UUID) ([]model.Poc, error) {
	type pocPlusJoin struct {
		ID              uuid.UUID       `bun:"id"`
		CreatedAt       time.Time       `bun:"created_at,nullzero"`
		UpdatedAt       time.Time       `bun:"updated_at,nullzero"`
		VulnerabilityID uuid.UUID       `bun:"vulnerability_id"`
		Items           []model.PocItem `bun:"items,type:jsonb"`
	}
	var rows []pocPlusJoin
	if err := idbFrom(ctx, pi.driver.db).NewSelect().
		TableExpr("poc AS p").
		ColumnExpr("DISTINCT p.id, p.created_at, p.updated_at, p.vulnerability_id, p.items").
		Join("JOIN poc_image pi ON pi.poc_id = p.id").
		Where("pi.file_reference_id = ?", imageID).
		Scan(ctx, &rows); err != nil {
		return nil, mapErr(err)
	}
	out := make([]model.Poc, len(rows))
	for i, r := range rows {
		p := model.Poc{
			VulnerabilityID: r.VulnerabilityID,
			Pocs:            r.Items,
		}
		p.ID = r.ID
		p.CreatedAt = r.CreatedAt
		p.UpdatedAt = r.UpdatedAt
		if p.Pocs == nil {
			p.Pocs = []model.PocItem{}
		}
		out[i] = p
	}
	return out, nil
}

func (pi *PocIndex) Clone(ctx context.Context, pocID, vulnerabilityID uuid.UUID) (uuid.UUID, error) {
	src, err := pi.GetByID(ctx, pocID)
	if err != nil {
		return uuid.Nil, err
	}
	if src == nil || src.ID == uuid.Nil {
		return uuid.Nil, store.ErrNotFound
	}

	newID, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	clonedItems := make([]model.PocItem, len(src.Pocs))
	for i, item := range src.Pocs {
		newItemID, err := uuid.NewRandom()
		if err != nil {
			return uuid.Nil, err
		}
		clonedItems[i] = item
		clonedItems[i].ID = newItemID
	}

	idb := idbFrom(ctx, pi.driver.db)
	if _, err := idb.NewInsert().
		Model(&dbPoc{
			ID:              newID,
			VulnerabilityID: vulnerabilityID,
			Items:           clonedItems,
		}).
		Exec(ctx); err != nil {
		return uuid.Nil, mapErr(err)
	}

	rows := make([]dbPocImage, 0, len(clonedItems))
	for _, item := range clonedItems {
		if item.ImageID == uuid.Nil {
			continue
		}
		rows = append(rows, dbPocImage{
			PocID:           newID,
			PocItemID:       item.ID,
			FileReferenceID: item.ImageID,
		})
	}
	if len(rows) > 0 {
		if _, err := idb.NewInsert().
			Model(&rows).
			Exec(ctx); err != nil {
			return uuid.Nil, mapErr(err)
		}
	}
	return newID, nil
}
