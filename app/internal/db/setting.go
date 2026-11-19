package db

import (
	"context"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type SettingIndex struct{ driver *Driver }

func (d *Driver) vulnerabilityFields(ctx context.Context) ([]model.VulnerabilityField, error) {
	var row dbSetting

	err := idbFrom(ctx, d.db).NewSelect().Model(&row).
		Column("vulnerability_fields").Where("id = ?", model.SettingID).Scan(ctx)

	return row.VulnerabilityFields, mapErr(err)
}

func (si *SettingIndex) Get(ctx context.Context) (*model.Setting, error) {
	var row dbSetting

	err := idbFrom(ctx, si.driver.db).NewSelect().
		Model(&row).
		Where("id = ?", model.SettingID).
		Scan(ctx)
	if err != nil {
		return nil, mapErr(err)
	}

	out := row.toModel()

	return &out, nil
}

func (si *SettingIndex) GetVulnerabilityFieldImpact(ctx context.Context, key string) (*model.VulnerabilityFieldImpact, error) {
	var impact model.VulnerabilityFieldImpact
	err := idbFrom(ctx, si.driver.db).NewRaw(`
		WITH affected_customers AS (
			SELECT id FROM customer WHERE ? = ANY(vulnerability_field_keys)
		), affected_vulnerabilities AS (
			SELECT assessment_id FROM vulnerability
			WHERE jsonb_exists(custom_fields, ?)
				OR customer_id IN (SELECT id FROM affected_customers)
		), affected_assessments AS (
			SELECT id FROM assessment WHERE customer_id IN (SELECT id FROM affected_customers)
			UNION SELECT assessment_id FROM affected_vulnerabilities
		)
		SELECT (SELECT count(*) FROM affected_vulnerabilities) AS vulnerabilities,
			(SELECT count(*) FROM affected_assessments) AS assessments
		FROM setting WHERE id = ? AND EXISTS (
			SELECT 1 FROM jsonb_array_elements(vulnerability_fields) AS field
			WHERE field->>'key' = ?
		)
	`, key, key, model.SettingID, key).Scan(ctx, &impact)
	if err != nil {
		return nil, mapErr(err)
	}

	return &impact, nil
}

func (si *SettingIndex) Update(ctx context.Context, setting *model.Setting) error {
	fields := setting.VulnerabilityFields
	q := idbFrom(ctx, si.driver.db).NewUpdate().
		Model((*dbSetting)(nil)).
		Set("max_image_size = ?", setting.MaxImageSize).
		Set("default_category_language = ?", setting.DefaultCategoryLanguage).
		Where("id = ?", model.SettingID)

	if fields == nil {
		_, err := q.Exec(ctx)

		return mapErr(err)
	}

	var previous dbSetting
	if err := idbFrom(ctx, si.driver.db).NewSelect().Model(&previous).
		Column("vulnerability_fields").Where("id = ?", model.SettingID).For("UPDATE").Scan(ctx); err != nil {
		return mapErr(err)
	}

	if err := model.ValidateVulnerabilityFields(fields, previous.VulnerabilityFields); err != nil {
		return err
	}

	retained := make(map[string]bool, len(fields))
	for _, field := range fields {
		retained[field.Key] = true
	}

	var removed []string
	for _, field := range previous.VulnerabilityFields {
		if !retained[field.Key] {
			removed = append(removed, field.Key)
		}
	}

	if len(removed) > 0 {
		keys := pgdialect.Array(removed)
		_, err := idbFrom(ctx, si.driver.db).NewRaw(`
			UPDATE customer SET vulnerability_field_keys = ARRAY(
				SELECT key FROM unnest(vulnerability_field_keys) WITH ORDINALITY AS fields(key, position)
				WHERE NOT (key = ANY(?::text[])) ORDER BY position
			) WHERE vulnerability_field_keys && ?::text[]
		`, keys, keys).Exec(ctx)
		if err != nil {
			return mapErr(err)
		}

		_, err = idbFrom(ctx, si.driver.db).NewRaw(`
			UPDATE vulnerability SET custom_fields = custom_fields - ?::text[]
			WHERE jsonb_exists_any(custom_fields, ?::text[])
		`, keys, keys).Exec(ctx)
		if err != nil {
			return mapErr(err)
		}
	}

	_, err := q.Set("vulnerability_fields = ?", fields).Exec(ctx)

	return mapErr(err)
}

func (si *SettingIndex) ValidateImageSize(ctx context.Context, size int64) error {
	s, err := si.Get(ctx)
	if err != nil {
		return err
	}

	if size > s.MaxImageSize {
		return store.ErrFileSizeTooLarge
	}

	return nil
}
