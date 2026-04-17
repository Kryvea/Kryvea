package mongo

import (
	"context"
	"crypto/md5"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	fileReferenceCollection = "file_reference"
)

type FileReference struct {
	Model    `bson:",inline"`
	File     bson.ObjectID `json:"file" bson:"file"`
	Checksum [16]byte      `json:"checksum" bson:"checksum"`
	MimeType string        `json:"mime_type" bson:"mime_type"`
	UsedBy   []uuid.UUID   `json:"used_by" bson:"used_by"`
}

type FileReferenceIndex struct {
	driver     *Driver
	collection *mongo.Collection
}

func (d *Driver) FileReference() *FileReferenceIndex {
	return &FileReferenceIndex{
		driver:     d,
		collection: d.database.Collection(fileReferenceCollection),
	}
}

func (fri *FileReferenceIndex) init() error {
	_, err := fri.collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "checksum", Value: 1},
			},
		},
	)
	return err
}

// Insert inserts a new file reference into the database if it does not already exist
// and uploads the file to the storage bucket.
//
// It returns the file reference ID, MIME type, and any error encountered.
//
// Requires transactional context to ensure data integrity.
func (i *FileReferenceIndex) Insert(ctx context.Context, data []byte) (uuid.UUID, string, error) {
	checksum := md5.Sum(data)
	reference, err := i.GetByChecksum(ctx, checksum)
	if err != nil && err != mongo.ErrNoDocuments {
		return uuid.Nil, "", err
	}
	if reference != nil {
		return reference.ID, reference.MimeType, nil
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, "", err
	}

	fileID, err := i.driver.File().Insert(ctx, data, id.String())
	if err != nil {
		return uuid.Nil, "", err
	}

	mime := mimetype.Detect(data).String()

	fileReference := FileReference{
		Model: Model{
			ID:        id,
			CreatedAt: time.Now(),
		},
		File:     fileID,
		Checksum: checksum,
		MimeType: mime,
		UsedBy:   []uuid.UUID{},
	}
	fileReference.Model.UpdatedAt = fileReference.Model.CreatedAt

	_, err = i.collection.InsertOne(ctx, fileReference)
	if err != nil {
		return uuid.Nil, "", err
	}

	return id, mime, nil
}

func (i *FileReferenceIndex) GetByID(ctx context.Context, id uuid.UUID) (*FileReference, error) {
	var fileReference FileReference
	err := i.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&fileReference)
	if err != nil {
		return nil, err
	}

	return &fileReference, nil
}

func (i *FileReferenceIndex) GetByChecksum(ctx context.Context, checksum [16]byte) (*FileReference, error) {
	var fileReference FileReference
	err := i.collection.FindOne(ctx, bson.M{"checksum": checksum}).Decode(&fileReference)
	if err != nil {
		return nil, err
	}

	return &fileReference, nil
}

func (i *FileReferenceIndex) ReadByID(ctx context.Context, id uuid.UUID) ([]byte, *FileReference, error) {
	fileReference, err := i.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	data, err := i.driver.File().GetByID(ctx, fileReference.File)
	if err != nil {
		return nil, nil, err
	}

	return data, fileReference, nil
}

func (i *FileReferenceIndex) PullUsedBy(ctx context.Context, id uuid.UUID, usedBy uuid.UUID) error {
	fileReference, err := i.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if len(fileReference.UsedBy) > 1 {
		_, err = i.collection.UpdateOne(ctx,
			bson.M{"_id": id},
			bson.M{"$pull": bson.M{"used_by": usedBy}},
		)
		return err
	}

	err = i.driver.File().Delete(ctx, fileReference.File)
	if err != nil {
		return err
	}

	_, err = i.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (i *FileReferenceIndex) AddToUsedBy(ctx context.Context, id uuid.UUID, usedBy uuid.UUID) error {
	if usedBy == uuid.Nil {
		return ErrUsedByIDRequired
	}

	_, err := i.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$addToSet": bson.M{"used_by": usedBy}},
	)
	return err
}
