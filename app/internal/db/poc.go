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

	newID, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	// vulnerability_id is unique: a single upsert either inserts a fresh row or
	// replaces the items of the existing one. RETURNING yields the winning id.
	row := &dbPoc{
		ID:              newID,
		VulnerabilityID: poc.VulnerabilityID,
		Items:           poc.Pocs,
	}
	if _, err := idb.NewInsert().
		Model(row).
		On("CONFLICT (vulnerability_id) DO UPDATE").
		Set("items = EXCLUDED.items").
		Returning("id").
		Exec(ctx); err != nil {
		return mapErr(err)
	}
	pocID := row.ID
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

func (pi *PocIndex) getOne(ctx context.Context, where string, arg uuid.UUID) (*model.Poc, error) {
	var row dbPoc
	if err := idbFrom(ctx, pi.driver.db).NewSelect().
		Model(&row).
		Where(where, arg).
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := row.toModel()
	return &out, nil
}

func (pi *PocIndex) GetByID(ctx context.Context, id uuid.UUID) (*model.Poc, error) {
	return pi.getOne(ctx, "id = ?", id)
}

func (pi *PocIndex) GetByVulnerabilityID(ctx context.Context, vulnerabilityID uuid.UUID) (*model.Poc, error) {
	return pi.getOne(ctx, "vulnerability_id = ?", vulnerabilityID)
}

func (pi *PocIndex) GetByImageID(ctx context.Context, imageID uuid.UUID) ([]model.Poc, error) {
	var rows []dbPoc
	if err := idbFrom(ctx, pi.driver.db).NewSelect().
		Model(&rows).
		Where("EXISTS (SELECT 1 FROM poc_image img WHERE img.poc_id = p.id AND img.file_reference_id = ?)", imageID).
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := make([]model.Poc, len(rows))
	for i := range rows {
		out[i] = rows[i].toModel()
	}
	return out, nil
}

// CloneByVulnerabilityID copies the poc of srcVulnerabilityID (if any) onto
// dstVulnerabilityID, assigning fresh poc and item IDs and re-deriving the
// poc_image rows. A vulnerability without a poc is a no-op.
func (pi *PocIndex) CloneByVulnerabilityID(ctx context.Context, srcVulnerabilityID, dstVulnerabilityID uuid.UUID) error {
	src, err := pi.GetByVulnerabilityID(ctx, srcVulnerabilityID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}
		return err
	}

	clonedItems := make([]model.PocItem, len(src.Pocs))
	for i, item := range src.Pocs {
		newItemID, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		clonedItems[i] = item
		clonedItems[i].ID = newItemID
	}

	idb := idbFrom(ctx, pi.driver.db)

	newPocID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	if _, err := idb.NewInsert().
		Model(&dbPoc{
			ID:              newPocID,
			VulnerabilityID: dstVulnerabilityID,
			Items:           clonedItems,
		}).
		Exec(ctx); err != nil {
		return mapErr(err)
	}

	rows := make([]dbPocImage, 0, len(clonedItems))
	for _, item := range clonedItems {
		if item.ImageID == uuid.Nil {
			continue
		}
		rows = append(rows, dbPocImage{
			PocID:           newPocID,
			PocItemID:       item.ID,
			FileReferenceID: item.ImageID,
		})
	}
	if len(rows) > 0 {
		if _, err := idb.NewInsert().
			Model(&rows).
			Exec(ctx); err != nil {
			return mapErr(err)
		}
	}
	return nil
}
