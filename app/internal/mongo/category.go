package mongo

import (
	"context"
	"regexp"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	categoryCollection = "category"

	SourceGeneric = "generic"
	SourceNessus  = "nessus"
	SourceBurp    = "burp"
)

type Category struct {
	Model              `bson:",inline"`
	Identifier         string            `json:"identifier" bson:"identifier"`
	Name               string            `json:"name" bson:"name"`
	Subcategory        string            `json:"subcategory,omitempty" bson:"subcategory"`
	GenericDescription map[string]string `json:"generic_description,omitempty" bson:"generic_description"`
	GenericRemediation map[string]string `json:"generic_remediation,omitempty" bson:"generic_remediation"`
	LanguagesOrder     []string          `json:"languages_order,omitempty" bson:"languages_order"`
	References         []string          `json:"references" bson:"references"`
	Source             string            `json:"source" bson:"source"`
}

type CategoryIndex struct {
	driver     *Driver
	collection *mongo.Collection
}

func (d *Driver) Category() *CategoryIndex {
	return &CategoryIndex{
		driver:     d,
		collection: d.database.Collection(categoryCollection),
	}
}

func (ci CategoryIndex) init() error {
	_, err := ci.collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "identifier", Value: 1},
				{Key: "name", Value: 1},
				{Key: "subcategory", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}

func (ci *CategoryIndex) Insert(ctx context.Context, category *Category) (uuid.UUID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	category.Model = Model{
		ID:        id,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	_, err = ci.collection.InsertOne(ctx, category)
	if err != nil {
		return uuid.Nil, err
	}

	return category.ID, err
}

func (ci *CategoryIndex) Upsert(ctx context.Context, category *Category, override bool) (uuid.UUID, error) {
	if !override {
		return ci.Insert(ctx, category)
	}

	id, isNew, err := ci.FirstOrInsert(ctx, category)
	if err == nil && isNew {
		return id, nil
	}

	if err != nil {
		return uuid.Nil, err
	}

	err = ci.Update(ctx, id, category)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (ci *CategoryIndex) FirstOrInsert(ctx context.Context, category *Category) (uuid.UUID, bool, error) {
	var existingCategory Assessment
	err := ci.collection.FindOne(ctx, bson.M{
		"identifier":  category.Identifier,
		"name":        category.Name,
		"subcategory": category.Subcategory,
	}).Decode(&existingCategory)
	if err == nil {
		return existingCategory.ID, false, nil
	}
	if err != mongo.ErrNoDocuments {
		return uuid.Nil, false, err
	}

	id, err := ci.Insert(ctx, category)
	return id, true, err
}

func (ci *CategoryIndex) Update(ctx context.Context, ID uuid.UUID, category *Category) error {
	if ID == ImmutableID {
		return ErrImmutableCategory
	}

	filter := bson.M{"_id": ID}

	update := bson.M{
		"$set": bson.M{
			"updated_at":          time.Now(),
			"identifier":          category.Identifier,
			"name":                category.Name,
			"subcategory":         category.Subcategory,
			"generic_description": category.GenericDescription,
			"generic_remediation": category.GenericRemediation,
			"languages_order":     category.LanguagesOrder,
			"references":          category.References,
			"source":              category.Source,
		},
	}

	_, err := ci.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete removes a category and reassigns associated vulnerabilities to the immutable category
//
// Requires transactional context to ensure data integrity
func (ci *CategoryIndex) Delete(ctx context.Context, ID uuid.UUID) error {
	if ID == ImmutableID {
		return ErrImmutableCategory
	}

	_, err := ci.collection.DeleteOne(ctx, bson.M{"_id": ID})
	if err != nil {
		return err
	}

	filter := bson.M{"category._id": ID}
	update := bson.M{
		"$set": bson.M{
			"category._id": ImmutableID,
		},
	}

	_, err = ci.driver.Vulnerability().collection.UpdateMany(ctx, filter, update)
	return err
}

func (ci *CategoryIndex) GetAll(ctx context.Context) ([]Category, error) {
	filter := bson.M{"_id": bson.M{"$ne": ImmutableID}}

	categories := []Category{}
	cursor, err := ci.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &categories)
	return categories, err
}

func (ci *CategoryIndex) GetByID(ctx context.Context, categoryID uuid.UUID) (*Category, error) {
	var category Category
	err := ci.collection.FindOne(ctx, bson.M{"_id": categoryID}).Decode(&category)
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (ci *CategoryIndex) Search(ctx context.Context, query string) ([]Category, error) {
	filter := bson.M{
		"_id": bson.M{"$ne": ImmutableID},
	}
	if query != "" {
		filter["$or"] = []bson.M{
			{"identifier": bson.M{"$regex": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}}},
			{"name": bson.M{"$regex": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}}},
			{"subcategory": bson.M{"$regex": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}}},
		}
	}

	cursor, err := ci.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	categories := []Category{}
	err = cursor.All(ctx, &categories)
	if err != nil {
		return nil, err
	}

	return categories, err
}
