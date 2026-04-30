package db

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
)

type FileReferenceIndex struct{ driver *Driver }

func (d *Driver) filePath(id uuid.UUID) string {
	s := id.String()
	return filepath.Join(d.filesDir, s[:2], s+".bin")
}

func (d *Driver) writeFile(id uuid.UUID, data []byte) error {
	path := d.filePath(id)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir file shard: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename file: %w", err)
	}
	return nil
}

func (d *Driver) readFile(id uuid.UUID) ([]byte, error) {
	return os.ReadFile(d.filePath(id))
}

func (d *Driver) deleteFile(id uuid.UUID) error {
	err := os.Remove(d.filePath(id))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (i *FileReferenceIndex) Insert(ctx context.Context, data []byte) (uuid.UUID, string, error) {
	checksum := md5.Sum(data)

	existing, err := i.GetByChecksum(ctx, checksum)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return uuid.Nil, "", err
	}
	if existing != nil {
		return existing.ID, existing.MimeType, nil
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, "", err
	}
	mime := mimetype.Detect(data).String()

	if err := i.driver.writeFile(id, data); err != nil {
		return uuid.Nil, "", err
	}

	row := &dbFileReference{
		ID:        id,
		Checksum:  checksum[:],
		MimeType:  mime,
		SizeBytes: int64(len(data)),
	}
	if _, err := idbFrom(ctx, i.driver.db).NewInsert().Model(row).Exec(ctx); err != nil {
		_ = i.driver.deleteFile(id)
		return uuid.Nil, "", mapErr(err)
	}
	return id, mime, nil
}

func (i *FileReferenceIndex) GetByID(ctx context.Context, id uuid.UUID) (*model.FileReference, error) {
	var row dbFileReference
	err := idbFrom(ctx, i.driver.db).NewSelect().
		Model(&row).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	out := row.toModel()

	usedBy, err := i.usedByOf(ctx, id)
	if err != nil {
		return nil, err
	}
	out.UsedBy = usedBy
	return &out, nil
}

func (i *FileReferenceIndex) usedByOf(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	idb := idbFrom(ctx, i.driver.db)
	var rows []uuid.UUID
	if err := idb.NewRaw(`
		SELECT id::uuid          FROM customer  WHERE logo_id           = ?
		UNION ALL
		SELECT id::uuid          FROM template  WHERE file_id           = ?
		UNION ALL
		SELECT poc_item_id::uuid FROM poc_image WHERE file_reference_id = ?
	`, id, id, id).Scan(ctx, &rows); err != nil {
		return nil, mapErr(err)
	}
	if rows == nil {
		rows = []uuid.UUID{}
	}
	return rows, nil
}

func (i *FileReferenceIndex) GetByChecksum(ctx context.Context, checksum [16]byte) (*model.FileReference, error) {
	var row dbFileReference
	err := idbFrom(ctx, i.driver.db).NewSelect().
		Model(&row).
		Where("checksum = ?", checksum[:]).
		Scan(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	out := row.toModel()
	return &out, nil
}

func (i *FileReferenceIndex) ReadByID(ctx context.Context, id uuid.UUID) ([]byte, *model.FileReference, error) {
	fr, err := i.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	data, err := i.driver.readFile(id)
	if err != nil {
		return nil, nil, fmt.Errorf("read file payload: %w", err)
	}
	return data, fr, nil
}

// GCFiles removes on-disk payloads whose file_reference row no longer exists.
// Returns the number of files removed.
func (i *FileReferenceIndex) GCFiles(ctx context.Context) (int, error) {
	root := i.driver.filesDir
	if root == "" {
		return 0, nil
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, fmt.Errorf("read files_dir: %w", err)
	}

	var ids []uuid.UUID
	if err := idbFrom(ctx, i.driver.db).NewSelect().
		Model((*dbFileReference)(nil)).
		Column("id").
		Scan(ctx, &ids); err != nil {
		return 0, mapErr(err)
	}
	known := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		known[id] = struct{}{}
	}

	removed := 0
	for _, shard := range entries {
		if !shard.IsDir() {
			continue
		}
		shardPath := filepath.Join(root, shard.Name())
		files, err := os.ReadDir(shardPath)
		if err != nil {
			return removed, fmt.Errorf("read shard %s: %w", shard.Name(), err)
		}
		for _, f := range files {
			name := f.Name()
			if !strings.HasSuffix(name, ".bin") {
				continue
			}
			id, err := uuid.Parse(strings.TrimSuffix(name, ".bin"))
			if err != nil {
				continue
			}
			if _, ok := known[id]; ok {
				continue
			}
			if err := i.driver.deleteFile(id); err != nil {
				i.driver.logger.Warn().Err(err).Str("id", id.String()).Msg("gc: failed to remove orphan file")
				continue
			}
			removed++
		}
	}
	return removed, nil
}
