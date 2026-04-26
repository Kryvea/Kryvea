package mongo

import (
	"context"
	"time"

	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	pocCollection = "poc"
)

type Poc struct {
	Model           `bson:",inline"`
	Pocs            []PocItem `json:"pocs" bson:"pocs"`
	VulnerabilityID uuid.UUID `json:"vulnerability_id" bson:"vulnerability_id"`
}

// TODO: should be reworked to allow unique relations with filereference
// maybe a simple ID parameter can work
type PocItem struct {
	Index               int               `json:"index" bson:"index"`
	Type                string            `json:"type" bson:"type"`
	Description         string            `json:"description" bson:"description"`
	URI                 string            `json:"uri,omitempty" bson:"uri,omitempty"`
	Request             string            `json:"request,omitempty" bson:"request,omitempty"`
	RequestHighlights   []HighlightedText `json:"request_highlights,omitempty" bson:"request_highlights,omitempty"`
	RequestHighlighted  []Highlighted     `json:"request_highlighted,omitempty" bson:"request_highlighted,omitempty"`
	Response            string            `json:"response,omitempty" bson:"response,omitempty"`
	ResponseHighlights  []HighlightedText `json:"response_highlights,omitempty" bson:"response_highlights,omitempty"`
	ResponseHighlighted []Highlighted     `json:"response_highlighted,omitempty" bson:"response_highlighted,omitempty"`
	ImageID             uuid.UUID         `json:"image_id,omitempty" bson:"image_id,omitempty"`
	ImageReference      string            `json:"image_reference,omitempty" bson:"image_reference,omitempty"`
	ImageFilename       string            `json:"image_filename,omitempty" bson:"image_filename,omitempty"`
	ImageMimeType       string            `json:"-" bson:"image_mimetype,omitempty"`
	ImageCaption        string            `json:"image_caption,omitempty" bson:"image_caption,omitempty"`
	TextLanguage        string            `json:"text_language,omitempty" bson:"text_language,omitempty"`
	TextData            string            `json:"text_data,omitempty" bson:"text_data,omitempty"`
	TextHighlights      []HighlightedText `json:"text_highlights,omitempty" bson:"text_highlights,omitempty"`
	TextHighlighted     []Highlighted     `json:"text_highlighted,omitempty" bson:"text_highlighted,omitempty"`
	StartingLineNumber  int               `json:"starting_line_number,omitempty" bson:"starting_line_number,omitempty"`
	// Only populated on report generation
	ImageData []byte `json:"-" bson:"-"`
}

type HighlightedText struct {
	Start           LineCol `json:"start" bson:"start"`
	End             LineCol `json:"end" bson:"end"`
	SelectedPreview string  `json:"selectionPreview" bson:"selection_preview"`
	Color           string  `json:"color" bson:"color"`
}

type LineCol struct {
	Line int `json:"line" bson:"line"`
	Col  int `json:"col" bson:"col"`
}

type Highlighted struct {
	Text  string `json:"text,omitempty"`
	Color string `json:"color,omitempty"`
}

type PocIndex struct {
	driver     *Driver
	collection *mongo.Collection
}

func (d *Driver) Poc() *PocIndex {
	return &PocIndex{
		driver:     d,
		collection: d.database.Collection(pocCollection),
	}
}

func (pi PocIndex) init() error {
	_, err := pi.collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "vulnerability_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}

// Upsert adds or updates a PoCs for a given vulnerability.
// It also manages the associated file references for images
// by deleting old references and adding new ones.
//
// Requires transactional context to ensure data integrity
func (pi *PocIndex) Upsert(ctx context.Context, poc *Poc) error {
	// retrieve existing POCs
	oldPoc, err := pi.GetByVulnerabilityID(ctx, poc.VulnerabilityID)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}

	// map new POC image IDs
	newImageIDs := make(map[uuid.UUID]struct{}, len(poc.Pocs))
	for i, newPocs := range poc.Pocs {
		if newPocs.ImageID == uuid.Nil {
			continue
		}

		newImageIDs[newPocs.ImageID] = struct{}{}
		poc.Pocs[i].ImageReference = util.CreateImageReference(newPocs.ImageMimeType, newPocs.ImageID)
	}

	// retrieve old POC images IDs that are not in the new POC
	oldImageIDs := make(map[uuid.UUID]struct{}, len(oldPoc.Pocs))
	for _, oldPocs := range oldPoc.Pocs {
		if oldPocs.ImageID == uuid.Nil {
			continue
		}

		if _, exists := newImageIDs[oldPocs.ImageID]; !exists {
			oldImageIDs[oldPocs.ImageID] = struct{}{}
		}
	}

	poc.UpdatedAt = time.Now()

	// Serialize without _id
	updateSet := bson.M{
		"pocs":             poc.Pocs,
		"vulnerability_id": poc.VulnerabilityID,
		"updated_at":       poc.UpdatedAt,
	}
	insertSet := bson.M{
		"_id":        uuid.New(),
		"created_at": time.Now(),
	}

	filter := bson.M{"vulnerability_id": poc.VulnerabilityID}
	upsert := bson.M{
		"$set":         updateSet,
		"$setOnInsert": insertSet,
	}

	_, err = pi.collection.UpdateOne(
		ctx,
		filter,
		upsert,
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return err
	}

	upsertedPoc, err := pi.GetByVulnerabilityID(ctx, poc.VulnerabilityID)
	if err != nil {
		return err
	}

	for _, pocItem := range poc.Pocs {
		if pocItem.ImageID == uuid.Nil {
			continue
		}

		err = pi.driver.FileReference().AddToUsedBy(ctx, pocItem.ImageID, upsertedPoc.ID)
		if err != nil {
			return err
		}
	}

	// delete old images that are not in the new POC
	for imageID := range oldImageIDs {
		if imageID != uuid.Nil {
			err = pi.driver.FileReference().PullUsedBy(ctx, imageID, oldPoc.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (pi *PocIndex) GetByID(ctx context.Context, ID uuid.UUID) (*Poc, error) {
	cursor, err := pi.collection.Find(ctx, bson.M{"_id": ID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	poc := &Poc{
		Pocs: []PocItem{},
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(poc); err != nil {
			return nil, err
		}
	}

	return poc, nil
}

func (pi *PocIndex) GetByVulnerabilityID(ctx context.Context, vulnerabilityID uuid.UUID) (*Poc, error) {
	filter := bson.M{"vulnerability_id": vulnerabilityID}
	opts := options.Find().SetLimit(1)

	cursor, err := pi.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	poc := &Poc{
		Pocs: []PocItem{},
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(poc); err != nil {
			return nil, err
		}
	}

	return poc, nil
}

func (pi *PocIndex) GetByImageID(ctx context.Context, imageID uuid.UUID) ([]Poc, error) {
	cursor, err := pi.collection.Find(ctx, bson.M{"pocs.image_id": imageID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	pocs := []Poc{}
	err = cursor.All(ctx, &pocs)
	if err != nil {
		return []Poc{}, err
	}

	return pocs, nil
}

// DeleteByVulnerabilityID removes PoCs associated with a given vulnerability ID.
// It also manages the associated file references for images.
//
// Requires transactional context to ensure data integrity
func (pi *PocIndex) DeleteByVulnerabilityID(ctx context.Context, vulnerabilityID uuid.UUID) error {
	poc, err := pi.GetByVulnerabilityID(ctx, vulnerabilityID)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}

	imageIDs := make(map[uuid.UUID]struct{}, len(poc.Pocs))
	for _, pocItem := range poc.Pocs {
		if _, exists := imageIDs[pocItem.ImageID]; exists {
			continue
		}

		err = pi.driver.FileReference().PullUsedBy(ctx, pocItem.ImageID, poc.ID)
		if err != nil {
			return err
		}

		imageIDs[pocItem.ImageID] = struct{}{}
	}

	_, err = pi.collection.DeleteOne(ctx, bson.M{"vulnerability_id": vulnerabilityID})
	return err
}

// DeleteManyByVulnerabilityID removes PoCs associated with given vulnerability IDs.
// It also manages the associated file references for images.
//
// Requires transactional context to ensure data integrity
func (pi *PocIndex) DeleteManyByVulnerabilityID(ctx context.Context, vulnerabilityIDs []uuid.UUID) error {
	for _, vulnerabilityID := range vulnerabilityIDs {
		poc, err := pi.GetByVulnerabilityID(ctx, vulnerabilityID)
		if err != nil && err != mongo.ErrNoDocuments {
			return err
		}

		imageIDs := make(map[uuid.UUID]struct{}, len(poc.Pocs))
		for _, pocItem := range poc.Pocs {
			if _, exists := imageIDs[pocItem.ImageID]; exists {
				continue
			}

			err = pi.driver.FileReference().PullUsedBy(ctx, pocItem.ImageID, poc.ID)
			if err != nil {
				return err
			}

			imageIDs[pocItem.ImageID] = struct{}{}
		}
	}

	filter := bson.M{"vulnerability_id": bson.M{"$in": vulnerabilityIDs}}
	_, err := pi.collection.DeleteMany(ctx, filter)
	return err
}

func (pi *PocIndex) Clone(ctx context.Context, pocID, vulnerabilityID uuid.UUID) (uuid.UUID, error) {
	poc, err := pi.GetByID(ctx, pocID)
	if err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	poc.ID = id
	poc.CreatedAt = time.Now()
	poc.UpdatedAt = poc.CreatedAt
	poc.VulnerabilityID = vulnerabilityID

	_, err = pi.collection.InsertOne(ctx, poc)
	if err != nil {
		return uuid.Nil, err
	}

	for _, pocItem := range poc.Pocs {
		if pocItem.ImageID == uuid.Nil {
			continue
		}

		err = pi.driver.FileReference().AddToUsedBy(ctx, pocItem.ImageID, poc.ID)
		if err != nil {
			return uuid.Nil, err
		}
	}
	return poc.ID, nil
}
