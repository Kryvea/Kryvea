package mongo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	settingCollection = "setting"
)

var (
	SettingID uuid.UUID = [16]byte{
		'K', 'R', 'Y', 'V',
		'E', 'A', '-', 'S',
		'E', 'T', 'T', 'I',
		'N', 'G', 'I', 'D',
	}
)

type Setting struct {
	Model                   `bson:",inline"`
	MaxImageSize            int64   `json:"-" bson:"max_image_size"`
	MaxImageSizeMB          float64 `json:"max_image_size" bson:"max_image_size_mb"` // Used for APIs
	DefaultCategoryLanguage string  `json:"default_category_language" bson:"default_category_language"`
}

type SettingIndex struct {
	driver     *Driver
	collection *mongo.Collection
}

func (d *Driver) Setting() *SettingIndex {
	return &SettingIndex{
		driver:     d,
		collection: d.database.Collection(settingCollection),
	}
}

func (ci SettingIndex) init() error {
	return nil
}

func (si *SettingIndex) Get(ctx context.Context) (*Setting, error) {
	setting := &Setting{}
	if err := si.collection.FindOne(ctx, bson.M{"_id": SettingID}).Decode(&setting); err != nil {
		return nil, err
	}
	return setting, nil
}

func (si *SettingIndex) Update(ctx context.Context, setting *Setting) error {
	filter := bson.M{"_id": SettingID}

	update := bson.M{
		"$set": bson.M{
			"updated_at":                time.Now(),
			"max_image_size":            setting.MaxImageSize,
			"max_image_size_mb":         setting.MaxImageSizeMB,
			"default_category_language": setting.DefaultCategoryLanguage,
		},
	}

	_, err := si.collection.UpdateOne(ctx, filter, update)
	return err
}

func (si *SettingIndex) ValidateImageSize(ctx context.Context, size int64) error {
	settings, err := si.Get(ctx)
	if err != nil {
		return err
	}

	if size > settings.MaxImageSize {
		return ErrFileSizeTooLarge
	}

	return nil
}
