package mongo

import (
	"bytes"
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type FileIndex struct {
	driver *Driver
}

func (d *Driver) File() *FileIndex {
	return &FileIndex{
		driver: d,
	}
}

func (i *FileIndex) init() error {
	id, err := i.Insert(context.Background(), []byte("init-file"), "init-file")
	if err != nil {
		return err
	}

	return i.driver.bucket.Delete(context.Background(), id)
}

func (i *FileIndex) Insert(ctx context.Context, data []byte, filename string) (bson.ObjectID, error) {
	id, err := i.driver.bucket.UploadFromStream(ctx, filename, bytes.NewReader(data))
	return id, err
}

func (i *FileIndex) GetByID(ctx context.Context, id bson.ObjectID) ([]byte, error) {
	var buf bytes.Buffer
	_, err := i.driver.bucket.DownloadToStream(ctx, id, &buf)
	return buf.Bytes(), err
}

func (i *FileIndex) Delete(ctx context.Context, id bson.ObjectID) error {
	return i.driver.bucket.Delete(ctx, id)
}
