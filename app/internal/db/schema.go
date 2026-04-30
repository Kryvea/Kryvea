package db

import (
	"context"
	"fmt"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/uptrace/bun"
)

func (d *Driver) applySchema(ctx context.Context) error {
	if _, err := d.db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto`); err != nil {
		return fmt.Errorf("create pgcrypto extension: %w", err)
	}
	if _, err := d.db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS pg_trgm`); err != nil {
		return fmt.Errorf("create pg_trgm extension: %w", err)
	}

	tables := []tableSpec{
		{model: (*dbSetting)(nil)},
		{model: (*dbFileReference)(nil)},
		{model: (*dbCustomer)(nil), fks: []string{
			`("logo_id") REFERENCES "file_reference" ("id") ON DELETE SET NULL`,
		}},
		{model: (*dbUser)(nil)},
		{model: (*dbUserCustomer)(nil), fks: []string{
			`("user_id") REFERENCES "users" ("id") ON DELETE CASCADE`,
			`("customer_id") REFERENCES "customer" ("id") ON DELETE CASCADE`,
		}},
		{model: (*dbCategory)(nil)},
		{model: (*dbTarget)(nil), fks: []string{
			`("customer_id") REFERENCES "customer" ("id") ON DELETE CASCADE`,
		}},
		{model: (*dbAssessment)(nil), fks: []string{
			`("customer_id") REFERENCES "customer" ("id") ON DELETE CASCADE`,
		}},
		{model: (*dbAssessmentTarget)(nil), fks: []string{
			`("assessment_id") REFERENCES "assessment" ("id") ON DELETE CASCADE`,
			`("target_id") REFERENCES "target" ("id") ON DELETE CASCADE`,
		}},
		{model: (*dbUserAssessment)(nil), fks: []string{
			`("user_id") REFERENCES "users" ("id") ON DELETE CASCADE`,
			`("assessment_id") REFERENCES "assessment" ("id") ON DELETE CASCADE`,
		}},
		{model: (*dbTemplate)(nil), fks: []string{
			`("file_id") REFERENCES "file_reference" ("id")`,
			`("customer_id") REFERENCES "customer" ("id") ON DELETE CASCADE`,
		}},
		{model: (*dbVulnerability)(nil), fks: []string{
			`("assessment_id") REFERENCES "assessment" ("id") ON DELETE CASCADE`,
			`("customer_id") REFERENCES "customer" ("id") ON DELETE CASCADE`,
			`("target_id") REFERENCES "target" ("id") ON DELETE SET DEFAULT`,
			`("user_id") REFERENCES "users" ("id") ON DELETE SET NULL`,
			`("category_id") REFERENCES "category" ("id") ON DELETE SET DEFAULT`,
		}},
		{model: (*dbPoc)(nil), fks: []string{
			`("vulnerability_id") REFERENCES "vulnerability" ("id") ON DELETE CASCADE`,
		}},
		{model: (*dbPocImage)(nil), fks: []string{
			`("poc_id") REFERENCES "poc" ("id") ON DELETE CASCADE`,
			`("file_reference_id") REFERENCES "file_reference" ("id") ON DELETE RESTRICT`,
		}},
	}
	for _, t := range tables {
		if err := createTable(ctx, d.db, t); err != nil {
			return err
		}
	}

	if _, err := d.db.ExecContext(ctx, `
		ALTER TABLE vulnerability ADD COLUMN IF NOT EXISTS search_text TEXT GENERATED ALWAYS AS (
			coalesce(detailed_title, '') || ' ' ||
			coalesce(status, '')         || ' ' ||
			coalesce(description, '')    || ' ' ||
			coalesce(remediation, '')
		) STORED
	`); err != nil {
		return fmt.Errorf("alter vulnerability.search_text: %w", err)
	}

	for _, stmt := range ddlStatements {
		if _, err := d.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("apply ddl: %w\n%s", err, stmt)
		}
	}

	if _, err := d.db.ExecContext(ctx, `
		UPDATE assessment a SET vulnerability_count = sub.cnt
		FROM (
			SELECT assessment_id, COUNT(*) AS cnt FROM vulnerability GROUP BY assessment_id
		) sub
		WHERE a.id = sub.assessment_id AND a.vulnerability_count <> sub.cnt
	`); err != nil {
		return fmt.Errorf("backfill vulnerability_count: %w", err)
	}

	if err := d.seedImmutableRows(ctx); err != nil {
		return err
	}

	d.logger.Debug().Msg("schema bootstrap complete")
	return nil
}

type tableSpec struct {
	model any
	fks   []string
}

func createTable(ctx context.Context, db *bun.DB, t tableSpec) error {
	q := db.NewCreateTable().Model(t.model).IfNotExists()
	for _, fk := range t.fks {
		q = q.ForeignKey(fk)
	}
	if _, err := q.Exec(ctx); err != nil {
		return fmt.Errorf("create table %T: %w", t.model, err)
	}
	return nil
}

var ddlStatements = []string{
	`ALTER TABLE setting DROP COLUMN IF EXISTS max_image_size_mb`,

	`CREATE INDEX IF NOT EXISTS idx_file_reference_checksum     ON file_reference (checksum)`,
	`CREATE INDEX IF NOT EXISTS idx_customer_name_trgm          ON customer USING gin (name gin_trgm_ops)`,
	`CREATE INDEX IF NOT EXISTS idx_user_username_trgm          ON users USING gin (username gin_trgm_ops)`,
	`CREATE INDEX IF NOT EXISTS idx_user_token                  ON users (token) WHERE token IS NOT NULL`,
	`CREATE INDEX IF NOT EXISTS idx_user_customer_customer      ON user_customer (customer_id)`,
	`CREATE INDEX IF NOT EXISTS idx_category_name_trgm          ON category USING gin (name gin_trgm_ops)`,
	`CREATE INDEX IF NOT EXISTS idx_category_identifier_trgm    ON category USING gin (identifier gin_trgm_ops)`,
	`CREATE INDEX IF NOT EXISTS idx_target_customer             ON target (customer_id)`,
	`CREATE INDEX IF NOT EXISTS idx_target_fqdn_trgm            ON target USING gin (fqdn gin_trgm_ops)`,
	`CREATE INDEX IF NOT EXISTS idx_target_ipv4_trgm            ON target USING gin (ipv4 gin_trgm_ops)`,
	`CREATE INDEX IF NOT EXISTS idx_assessment_customer         ON assessment (customer_id)`,
	`CREATE INDEX IF NOT EXISTS idx_assessment_name_trgm        ON assessment USING gin (name gin_trgm_ops)`,
	`CREATE INDEX IF NOT EXISTS idx_assessment_target_target    ON assessment_target (target_id)`,
	`CREATE INDEX IF NOT EXISTS idx_user_assessment_assessment  ON user_assessment (assessment_id)`,
	`CREATE INDEX IF NOT EXISTS idx_template_customer           ON template (customer_id)`,
	`CREATE INDEX IF NOT EXISTS idx_template_file               ON template (file_id)`,
	`CREATE INDEX IF NOT EXISTS idx_vuln_assessment             ON vulnerability (assessment_id)`,
	`CREATE INDEX IF NOT EXISTS idx_vuln_customer               ON vulnerability (customer_id)`,
	`CREATE INDEX IF NOT EXISTS idx_vuln_target                 ON vulnerability (target_id)`,
	`CREATE INDEX IF NOT EXISTS idx_vuln_user                   ON vulnerability (user_id)`,
	`CREATE INDEX IF NOT EXISTS idx_vuln_category               ON vulnerability (category_id)`,
	`CREATE INDEX IF NOT EXISTS idx_vuln_updated_at             ON vulnerability (updated_at DESC)`,
	`CREATE INDEX IF NOT EXISTS idx_vuln_assessment_updated     ON vulnerability (assessment_id, updated_at DESC)`,
	`CREATE INDEX IF NOT EXISTS idx_vuln_search_text_trgm       ON vulnerability USING gin (search_text gin_trgm_ops)`,
	`CREATE INDEX IF NOT EXISTS idx_poc_image_file_reference    ON poc_image (file_reference_id)`,

	`CREATE OR REPLACE FUNCTION kryvea_set_updated_at() RETURNS TRIGGER AS $func$
		BEGIN
			NEW.updated_at = now();
			RETURN NEW;
		END;
	$func$ LANGUAGE plpgsql`,

	`DO $do$
		DECLARE
			t            text;
			trigger_name text;
		BEGIN
			FOR t IN SELECT unnest(ARRAY[
				'setting','file_reference','customer','users','category','target',
				'assessment','template','vulnerability','poc'
			])
			LOOP
				trigger_name := 'trg_' || t || '_updated_at';
				EXECUTE format('DROP TRIGGER IF EXISTS %I ON %I', trigger_name, t);
				EXECUTE format(
					'CREATE TRIGGER %I BEFORE UPDATE ON %I '
					'FOR EACH ROW EXECUTE FUNCTION kryvea_set_updated_at()',
					trigger_name, t
				);
			END LOOP;
		END;
	$do$`,

	`CREATE OR REPLACE FUNCTION kryvea_vuln_count_sync() RETURNS TRIGGER AS $func$
		BEGIN
			IF TG_OP = 'INSERT' THEN
				UPDATE assessment SET vulnerability_count = vulnerability_count + 1
					WHERE id = NEW.assessment_id;
				RETURN NEW;
			ELSIF TG_OP = 'DELETE' THEN
				UPDATE assessment SET vulnerability_count = vulnerability_count - 1
					WHERE id = OLD.assessment_id;
				RETURN OLD;
			ELSIF TG_OP = 'UPDATE' AND NEW.assessment_id <> OLD.assessment_id THEN
				UPDATE assessment SET vulnerability_count = vulnerability_count - 1
					WHERE id = OLD.assessment_id;
				UPDATE assessment SET vulnerability_count = vulnerability_count + 1
					WHERE id = NEW.assessment_id;
				RETURN NEW;
			END IF;
			RETURN NEW;
		END;
	$func$ LANGUAGE plpgsql`,

	`DROP TRIGGER IF EXISTS trg_vulnerability_count ON vulnerability`,
	`CREATE TRIGGER trg_vulnerability_count
		AFTER INSERT OR DELETE OR UPDATE OF assessment_id ON vulnerability
		FOR EACH ROW EXECUTE FUNCTION kryvea_vuln_count_sync()`,

	fmt.Sprintf(`ALTER TABLE vulnerability ALTER COLUMN target_id SET DEFAULT '%s'`, model.ImmutableID),
	fmt.Sprintf(`ALTER TABLE vulnerability ALTER COLUMN category_id SET DEFAULT '%s'`, model.ImmutableID),

	`UPDATE poc SET items = (
		SELECT jsonb_agg(
			CASE
				WHEN item ? 'id' AND (item->>'id') <> '' THEN item
				ELSE item || jsonb_build_object('id', gen_random_uuid())
			END
			ORDER BY ord
		)
		FROM jsonb_array_elements(items) WITH ORDINALITY AS arr(item, ord)
	)
	WHERE EXISTS (
		SELECT 1 FROM jsonb_array_elements(items) AS x(item)
		WHERE NOT (x.item ? 'id') OR (x.item->>'id') = ''
	)`,

	`INSERT INTO poc_image (poc_id, poc_item_id, file_reference_id)
	SELECT p.id, (item->>'id')::uuid, (item->>'image_id')::uuid
	FROM poc p, jsonb_array_elements(p.items) item
	WHERE (item->>'image_id') IS NOT NULL
	  AND (item->>'image_id') <> ''
	  AND (item->>'image_id') <> '00000000-0000-0000-0000-000000000000'
	ON CONFLICT (poc_id, poc_item_id) DO NOTHING`,

	`DROP TABLE IF EXISTS poc_image_usage`,
	`DROP TABLE IF EXISTS file_reference_usage`,

	`CREATE OR REPLACE FUNCTION kryvea_gc_file_reference() RETURNS TRIGGER AS $func$
		DECLARE
			fr_id uuid;
		BEGIN
			IF TG_TABLE_NAME = 'poc_image' THEN
				fr_id := OLD.file_reference_id;
				IF TG_OP = 'UPDATE' AND OLD.file_reference_id IS NOT DISTINCT FROM NEW.file_reference_id THEN
					RETURN NULL;
				END IF;
			ELSIF TG_TABLE_NAME = 'customer' THEN
				fr_id := OLD.logo_id;
				IF TG_OP = 'UPDATE' AND OLD.logo_id IS NOT DISTINCT FROM NEW.logo_id THEN
					RETURN NULL;
				END IF;
			ELSIF TG_TABLE_NAME = 'template' THEN
				fr_id := OLD.file_id;
				IF TG_OP = 'UPDATE' AND OLD.file_id IS NOT DISTINCT FROM NEW.file_id THEN
					RETURN NULL;
				END IF;
			END IF;

			IF fr_id IS NULL THEN RETURN NULL; END IF;

			IF NOT EXISTS (
				SELECT 1 FROM customer  WHERE logo_id           = fr_id
				UNION ALL
				SELECT 1 FROM template  WHERE file_id           = fr_id
				UNION ALL
				SELECT 1 FROM poc_image WHERE file_reference_id = fr_id
			) THEN
				DELETE FROM file_reference WHERE id = fr_id;
			END IF;
			RETURN NULL;
		END;
	$func$ LANGUAGE plpgsql`,

	`DROP TRIGGER IF EXISTS trg_customer_file_gc ON customer`,
	`CREATE TRIGGER trg_customer_file_gc
		AFTER UPDATE OF logo_id OR DELETE ON customer
		FOR EACH ROW EXECUTE FUNCTION kryvea_gc_file_reference()`,

	`DROP TRIGGER IF EXISTS trg_template_file_gc ON template`,
	`CREATE TRIGGER trg_template_file_gc
		AFTER UPDATE OF file_id OR DELETE ON template
		FOR EACH ROW EXECUTE FUNCTION kryvea_gc_file_reference()`,

	`DROP TRIGGER IF EXISTS trg_poc_image_file_gc ON poc_image`,
	`CREATE TRIGGER trg_poc_image_file_gc
		AFTER UPDATE OF file_reference_id OR DELETE ON poc_image
		FOR EACH ROW EXECUTE FUNCTION kryvea_gc_file_reference()`,
}

func (d *Driver) seedImmutableRows(ctx context.Context) error {
	if _, err := d.db.NewInsert().Model(&dbSetting{ID: model.SettingID}).
		On("CONFLICT (id) DO NOTHING").Exec(ctx); err != nil {
		return fmt.Errorf("seed setting: %w", err)
	}

	if _, err := d.db.NewInsert().Model(&dbCategory{
		ID:                 model.ImmutableID,
		Identifier:         "KRYVEA",
		Name:               "DELETED-CATEGORY",
		Source:             model.SourceGeneric,
		GenericDescription: map[string]string{"en": "The original category for this vulnerability has been deleted, please select a new one"},
		GenericRemediation: map[string]string{},
		LanguagesOrder:     []string{},
		References:         []string{},
	}).On("CONFLICT (id) DO NOTHING").Exec(ctx); err != nil {
		return fmt.Errorf("seed immutable category: %w", err)
	}

	if _, err := d.db.NewInsert().Model(&dbTarget{
		ID: model.ImmutableID,
	}).On("CONFLICT (id) DO NOTHING").Exec(ctx); err != nil {
		return fmt.Errorf("seed immutable target: %w", err)
	}

	return nil
}
