-- Initial schema for an empty application database
SET LOCAL search_path TO public;
SET LOCAL lock_timeout TO '30s';

--bun:split

CREATE EXTENSION IF NOT EXISTS pgcrypto;

--bun:split

CREATE EXTENSION IF NOT EXISTS pg_trgm;

--bun:split

CREATE TABLE "setting" (
    "id" uuid NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "max_image_size" BIGINT NOT NULL DEFAULT 5242880,
    "default_category_language" VARCHAR NOT NULL DEFAULT 'en',
    PRIMARY KEY ("id"));

--bun:split

CREATE TABLE "file_reference" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "checksum" BYTEA NOT NULL,
    "mime_type" VARCHAR NOT NULL,
    PRIMARY KEY ("id"),
    UNIQUE ("checksum"));

--bun:split

CREATE TABLE "customer" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "name" VARCHAR NOT NULL,
    "language" VARCHAR NOT NULL DEFAULT '',
    "logo_id" uuid,
    "logo_mime_type" VARCHAR NOT NULL DEFAULT '',
    "logo_reference" VARCHAR NOT NULL DEFAULT '',
    PRIMARY KEY ("id"),
    UNIQUE ("name"),
    FOREIGN KEY ("logo_id") REFERENCES "file_reference" ("id") ON DELETE SET NULL);

--bun:split

CREATE TABLE "users" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "disabled_at" TIMESTAMPTZ,
    "username" VARCHAR NOT NULL,
    "password" BYTEA NOT NULL,
    "password_expiry" TIMESTAMPTZ NOT NULL,
    "token" BYTEA,
    "token_expiry" TIMESTAMPTZ,
    "role" VARCHAR NOT NULL,
    PRIMARY KEY ("id"),
    UNIQUE ("username"));

--bun:split

CREATE TABLE "user_customer" (
    "user_id" uuid NOT NULL,
    "customer_id" uuid NOT NULL,
    PRIMARY KEY ("user_id",
    "customer_id"),
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("customer_id") REFERENCES "customer" ("id") ON DELETE CASCADE);

--bun:split

CREATE TABLE "category" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "identifier" VARCHAR NOT NULL DEFAULT '',
    "name" VARCHAR NOT NULL DEFAULT '',
    "subcategory" VARCHAR NOT NULL DEFAULT '',
    "generic_description" jsonb NOT NULL DEFAULT '{}',
    "generic_remediation" jsonb NOT NULL DEFAULT '{}',
    "languages_order" VARCHAR[] NOT NULL DEFAULT '{}',
    "refs" VARCHAR[] NOT NULL DEFAULT '{}',
    "source" VARCHAR NOT NULL DEFAULT 'generic',
    PRIMARY KEY ("id"), CONSTRAINT "category_natural_key" UNIQUE ("identifier",
    "name",
    "subcategory"));

--bun:split

CREATE TABLE "target" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "ipv4" VARCHAR NOT NULL DEFAULT '',
    "ipv6" VARCHAR NOT NULL DEFAULT '',
    "fqdn" VARCHAR NOT NULL DEFAULT '',
    "tag" VARCHAR NOT NULL DEFAULT '',
    "protocol" VARCHAR NOT NULL DEFAULT '',
    "port" BIGINT NOT NULL DEFAULT 0,
    "customer_id" uuid,
    PRIMARY KEY ("id"), CONSTRAINT "target_natural_key" UNIQUE ("ipv4",
    "ipv6",
    "fqdn",
    "tag",
    "port",
    "protocol",
    "customer_id"),
    FOREIGN KEY ("customer_id") REFERENCES "customer" ("id") ON DELETE CASCADE);

--bun:split

CREATE TABLE "assessment" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "customer_id" uuid NOT NULL,
    "name" VARCHAR NOT NULL,
    "language" VARCHAR NOT NULL,
    "start_date_time" TIMESTAMPTZ,
    "end_date_time" TIMESTAMPTZ,
    "kickoff_date_time" TIMESTAMPTZ,
    "status" VARCHAR NOT NULL DEFAULT '',
    "type_short" VARCHAR NOT NULL DEFAULT '',
    "type_full" VARCHAR NOT NULL DEFAULT '',
    "cvss_versions" jsonb NOT NULL DEFAULT '{}',
    "environment" VARCHAR NOT NULL DEFAULT '',
    "testing_type" VARCHAR NOT NULL DEFAULT '',
    "osstmm_vector" VARCHAR NOT NULL DEFAULT '',
    "vulnerability_count" BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY ("id"), CONSTRAINT "assessment_natural_key" UNIQUE ("customer_id",
    "name",
    "language"),
    FOREIGN KEY ("customer_id") REFERENCES "customer" ("id") ON DELETE CASCADE);

--bun:split

CREATE TABLE "assessment_target" (
    "assessment_id" uuid NOT NULL,
    "target_id" uuid NOT NULL,
    PRIMARY KEY ("assessment_id",
    "target_id"),
    FOREIGN KEY ("assessment_id") REFERENCES "assessment" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("target_id") REFERENCES "target" ("id") ON DELETE CASCADE);

--bun:split

CREATE TABLE "user_assessment" (
    "user_id" uuid NOT NULL,
    "assessment_id" uuid NOT NULL,
    PRIMARY KEY ("user_id",
    "assessment_id"),
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("assessment_id") REFERENCES "assessment" ("id") ON DELETE CASCADE);

--bun:split

CREATE TABLE "template" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "name" VARCHAR NOT NULL,
    "filename" VARCHAR NOT NULL DEFAULT '',
    "language" VARCHAR NOT NULL DEFAULT '',
    "template_type" VARCHAR NOT NULL,
    "mime_type" VARCHAR NOT NULL,
    "identifier" VARCHAR NOT NULL DEFAULT '',
    "file_id" uuid NOT NULL,
    "customer_id" uuid,
    PRIMARY KEY ("id"), CONSTRAINT "template_natural_key" UNIQUE ("name",
    "filename",
    "language",
    "template_type",
    "identifier"),
    FOREIGN KEY ("file_id") REFERENCES "file_reference" ("id"),
    FOREIGN KEY ("customer_id") REFERENCES "customer" ("id") ON DELETE CASCADE);

--bun:split

CREATE TABLE "vulnerability" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "assessment_id" uuid NOT NULL,
    "customer_id" uuid NOT NULL,
    "target_id" uuid NOT NULL,
    "user_id" uuid,
    "category_id" uuid NOT NULL,
    "detailed_title" VARCHAR NOT NULL DEFAULT '',
    "status" VARCHAR NOT NULL DEFAULT '',
    "cvssv2_vector" VARCHAR NOT NULL DEFAULT '',
    "cvssv2_score" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "cvssv3_vector" VARCHAR NOT NULL DEFAULT '',
    "cvssv3_score" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "cvssv31_vector" VARCHAR NOT NULL DEFAULT '',
    "cvssv31_score" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "cvssv4_vector" VARCHAR NOT NULL DEFAULT '',
    "cvssv4_score" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "refs" VARCHAR[] NOT NULL DEFAULT '{}',
    "description" VARCHAR NOT NULL DEFAULT '',
    "remediation" VARCHAR NOT NULL DEFAULT '',
    "generic_description_enabled" BOOLEAN NOT NULL DEFAULT true,
    "generic_remediation_enabled" BOOLEAN NOT NULL DEFAULT false,
    "generic_remediation_text" VARCHAR NOT NULL DEFAULT '',
    PRIMARY KEY ("id"),
    FOREIGN KEY ("assessment_id") REFERENCES "assessment" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("customer_id") REFERENCES "customer" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("target_id") REFERENCES "target" ("id") ON DELETE SET DEFAULT,
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE SET NULL,
    FOREIGN KEY ("category_id") REFERENCES "category" ("id") ON DELETE SET DEFAULT);

--bun:split

CREATE TABLE "poc" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "vulnerability_id" uuid NOT NULL,
    "items" jsonb NOT NULL DEFAULT '[]',
    PRIMARY KEY ("id"),
    UNIQUE ("vulnerability_id"),
    FOREIGN KEY ("vulnerability_id") REFERENCES "vulnerability" ("id") ON DELETE CASCADE);

--bun:split

CREATE TABLE "poc_image" (
    "poc_id" uuid NOT NULL,
    "poc_item_id" uuid NOT NULL,
    "file_reference_id" uuid NOT NULL,
    PRIMARY KEY ("poc_id",
    "poc_item_id"),
    FOREIGN KEY ("poc_id") REFERENCES "poc" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("file_reference_id") REFERENCES "file_reference" ("id") ON DELETE RESTRICT);

--bun:split

ALTER TABLE vulnerability ADD COLUMN search_text TEXT GENERATED ALWAYS AS (
			coalesce(detailed_title, '') || ' ' ||
			coalesce(status, '')         || ' ' ||
			coalesce(description, '')    || ' ' ||
			coalesce(remediation, '')
		) STORED;

--bun:split

CREATE INDEX idx_customer_name_trgm          ON customer USING gin (name gin_trgm_ops);

--bun:split

CREATE INDEX idx_user_username_trgm          ON users USING gin (username gin_trgm_ops);

--bun:split

CREATE INDEX idx_user_token                  ON users (token) WHERE token IS NOT NULL;

--bun:split

CREATE INDEX idx_user_customer_customer      ON user_customer (customer_id);

--bun:split

CREATE INDEX idx_category_name_trgm          ON category USING gin (name gin_trgm_ops);

--bun:split

CREATE INDEX idx_category_identifier_trgm    ON category USING gin (identifier gin_trgm_ops);

--bun:split

CREATE INDEX idx_target_customer             ON target (customer_id);

--bun:split

CREATE INDEX idx_target_fqdn_trgm            ON target USING gin (fqdn gin_trgm_ops);

--bun:split

CREATE INDEX idx_target_ipv4_trgm            ON target USING gin (ipv4 gin_trgm_ops);

--bun:split

CREATE INDEX idx_assessment_customer         ON assessment (customer_id);

--bun:split

CREATE INDEX idx_assessment_name_trgm        ON assessment USING gin (name gin_trgm_ops);

--bun:split

CREATE INDEX idx_assessment_target_target    ON assessment_target (target_id);

--bun:split

CREATE INDEX idx_user_assessment_assessment  ON user_assessment (assessment_id);

--bun:split

CREATE INDEX idx_template_customer           ON template (customer_id);

--bun:split

CREATE INDEX idx_template_file               ON template (file_id);

--bun:split

CREATE INDEX idx_vuln_assessment             ON vulnerability (assessment_id);

--bun:split

CREATE INDEX idx_vuln_customer               ON vulnerability (customer_id);

--bun:split

CREATE INDEX idx_vuln_target                 ON vulnerability (target_id);

--bun:split

CREATE INDEX idx_vuln_user                   ON vulnerability (user_id);

--bun:split

CREATE INDEX idx_vuln_category               ON vulnerability (category_id);

--bun:split

CREATE INDEX idx_vuln_updated_at             ON vulnerability (updated_at DESC);

--bun:split

CREATE INDEX idx_vuln_assessment_updated     ON vulnerability (assessment_id, updated_at DESC);

--bun:split

CREATE INDEX idx_vuln_search_text_trgm       ON vulnerability USING gin (search_text gin_trgm_ops);

--bun:split

CREATE INDEX idx_poc_image_file_reference    ON poc_image (file_reference_id);

--bun:split

CREATE FUNCTION kryvea_set_updated_at() RETURNS TRIGGER AS $func$
		BEGIN
			NEW.updated_at = now();
			RETURN NEW;
		END;
	$func$ LANGUAGE plpgsql;

--bun:split

DO $do$
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
				EXECUTE format(
					'CREATE TRIGGER %I BEFORE UPDATE ON %I '
					'FOR EACH ROW EXECUTE FUNCTION kryvea_set_updated_at()',
					trigger_name, t
				);
			END LOOP;
		END;
	$do$;

--bun:split

CREATE FUNCTION kryvea_vuln_count_sync() RETURNS TRIGGER AS $func$
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
	$func$ LANGUAGE plpgsql;

--bun:split

CREATE TRIGGER trg_vulnerability_count
		AFTER INSERT OR DELETE OR UPDATE OF assessment_id ON vulnerability
		FOR EACH ROW EXECUTE FUNCTION kryvea_vuln_count_sync();

--bun:split

ALTER TABLE vulnerability ALTER COLUMN target_id SET DEFAULT '4b525956-4541-2d49-4d4d-555441424c45';

--bun:split

ALTER TABLE vulnerability ALTER COLUMN category_id SET DEFAULT '4b525956-4541-2d49-4d4d-555441424c45';

--bun:split

CREATE FUNCTION kryvea_gc_file_reference() RETURNS TRIGGER AS $func$
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
	$func$ LANGUAGE plpgsql;

--bun:split

CREATE TRIGGER trg_customer_file_gc
		AFTER UPDATE OF logo_id OR DELETE ON customer
		FOR EACH ROW EXECUTE FUNCTION kryvea_gc_file_reference();

--bun:split

CREATE TRIGGER trg_template_file_gc
		AFTER UPDATE OF file_id OR DELETE ON template
		FOR EACH ROW EXECUTE FUNCTION kryvea_gc_file_reference();

--bun:split

CREATE TRIGGER trg_poc_image_file_gc
		AFTER UPDATE OF file_reference_id OR DELETE ON poc_image
		FOR EACH ROW EXECUTE FUNCTION kryvea_gc_file_reference();

--bun:split

INSERT INTO setting ("id")
VALUES ('4b525956-4541-2d53-4554-54494e474944');

--bun:split

INSERT INTO category ("id", "identifier", "name", "generic_description")
VALUES (
    '4b525956-4541-2d49-4d4d-555441424c45',
    'KRYVEA',
    'DELETED-CATEGORY',
    '{"en":"The original category for this vulnerability has been deleted, please select a new one"}'
);

--bun:split

INSERT INTO target ("id", "fqdn")
VALUES ('4b525956-4541-2d49-4d4d-555441424c45', 'deleted.target');
