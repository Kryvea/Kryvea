package mongo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	templateCollection = "template"
)

var (
	TemplateTypeXlsx           = "xlsx"
	TemplateTypeDocx           = "docx"
	TemplateTypeZip            = "generic-zip"
	XlsxMimeType               = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	DocxMimeType               = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	ZipMimeType                = "application/zip"
	SupportedTemplateMimeTypes = map[string]string{
		XlsxMimeType: TemplateTypeXlsx,
		DocxMimeType: TemplateTypeDocx,
		ZipMimeType:  TemplateTypeZip,
	}
)

type Template struct {
	Model        `bson:",inline"`
	Name         string    `json:"name" bson:"name"`
	Filename     string    `json:"filename,omitempty" bson:"filename"`
	Language     string    `json:"language,omitempty" bson:"language"`
	TemplateType string    `json:"template_type" bson:"template_type"`
	MimeType     string    `json:"-" bson:"mime_type"`
	Identifier   string    `json:"identifier,omitempty" bson:"identifier"`
	FileID       uuid.UUID `json:"file_id,omitempty" bson:"file_id"`
	Customer     *Customer `json:"customer,omitempty" bson:"customer"`
}

type TemplateIndex struct {
	driver     *Driver
	collection *mongo.Collection
}

func (d *Driver) Template() *TemplateIndex {
	return &TemplateIndex{
		driver:     d,
		collection: d.database.Collection(templateCollection),
	}
}

func (ti TemplateIndex) init() error {
	_, err := ti.collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "name", Value: 1},
				{Key: "filename", Value: 1},
				{Key: "language", Value: 1},
				{Key: "template_type", Value: 1},
				{Key: "identifier", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}

// Insert adds a new template to the database including file reference
//
// Requires transactional context to ensure data integrity
func (ti *TemplateIndex) Insert(ctx context.Context, template *Template) (uuid.UUID, error) {
	if template.FileID == uuid.Nil {
		return uuid.Nil, ErrTemplateFileIDRequired
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	template.Model = Model{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = ti.collection.InsertOne(ctx, template)
	if err != nil {
		return uuid.Nil, err
	}

	err = ti.driver.FileReference().AddToUsedBy(ctx, template.FileID, template.ID)
	if err != nil {
		return uuid.Nil, err
	}

	return template.ID, err
}

func (ti *TemplateIndex) GetByID(ctx context.Context, id uuid.UUID) (*Template, error) {
	var template Template
	err := ti.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&template)
	if err != nil {
		return nil, err
	}

	err = ti.hydrate(ctx, &template)
	if err != nil {
		return nil, err
	}

	return &template, nil
}

func (ti *TemplateIndex) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]Template, error) {
	filter := bson.M{"customer._id": customerID}
	opts := options.Find().SetSort(bson.D{
		{Key: "created_at", Value: -1},
	})

	cursor, err := ti.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	templates := []Template{}
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, err
	}

	for i := range templates {
		err = ti.hydrate(ctx, &templates[i])
		if err != nil {
			return nil, err
		}
	}

	return templates, nil
}

func (ti *TemplateIndex) GetByCustomerIDForHydrate(ctx context.Context, customerID uuid.UUID) ([]Template, error) {
	filter := bson.M{"customer._id": customerID}
	opts := options.Find().SetSort(bson.D{
		{Key: "created_at", Value: -1},
	}).SetProjection(bson.M{
		"customer": 0,
	})

	cursor, err := ti.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	templates := []Template{}
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, err
	}

	return templates, nil
}

func (ti *TemplateIndex) GetByFileID(ctx context.Context, fileID uuid.UUID) (*Template, error) {
	var template Template
	err := ti.collection.FindOne(ctx, bson.D{{Key: "file_id", Value: fileID}}).Decode(&template)
	if err != nil {
		return nil, err
	}

	return &template, nil
}

func (ti *TemplateIndex) GetAll(ctx context.Context) ([]Template, error) {
	filter := bson.M{}
	cursor, err := ti.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	templates := []Template{}
	err = cursor.All(ctx, &templates)
	if err != nil {
		return nil, err
	}

	for i := range templates {
		err = ti.hydrate(ctx, &templates[i])
		if err != nil {
			return nil, err
		}
	}

	return templates, nil
}

// Delete removes a template and its file reference
//
// Requires transactional context to ensure data integrity
func (ti *TemplateIndex) Delete(ctx context.Context, id uuid.UUID) error {
	template, err := ti.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// remove template id from FileReference
	err = ti.driver.FileReference().PullUsedBy(ctx, template.FileID, template.ID)
	if err != nil {
		return err
	}

	_, err = ti.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	return err
}

// hydrate fills in the nested fields for a Template
func (ti *TemplateIndex) hydrate(ctx context.Context, template *Template) error {
	// customer is optional
	if template.Customer.ID != uuid.Nil {
		customer, err := ti.driver.Customer().GetByIDForHydrate(ctx, template.Customer.ID)
		if err != nil {
			return err
		}

		template.Customer = customer
	}

	return nil
}
