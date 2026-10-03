package mongo

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	customerCollection = "customer"
)

type Customer struct {
	Model         `bson:",inline"`
	Name          string     `json:"name" bson:"name"`
	Language      string     `json:"language" bson:"language"`
	LogoID        uuid.UUID  `json:"logo_id" bson:"logo_id"`
	LogoMimeType  string     `json:"-" bson:"logo_mime_type"`
	LogoReference string     `json:"logo_reference" bson:"logo_reference"`
	Templates     []Template `json:"templates" bson:"templates"`

	LogoData []byte `json:"-" bson:"-"`
}

type CustomerIndex struct {
	driver     *Driver
	collection *mongo.Collection
}

func (d *Driver) Customer() *CustomerIndex {
	return &CustomerIndex{
		driver:     d,
		collection: d.database.Collection(customerCollection),
	}
}

func (ci CustomerIndex) init() error {
	_, err := ci.collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "name", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}

// Insert adds a new customer to the database including logo reference
//
// Requires transactional context to ensure data integrity
func (ci *CustomerIndex) Insert(ctx context.Context, customer *Customer) (uuid.UUID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	if customer.LogoID != uuid.Nil {
		customer.LogoReference = util.CreateImageReference(customer.LogoMimeType, customer.LogoID)
	}

	customer.Model = Model{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = ci.collection.InsertOne(ctx, customer)
	if err != nil {
		return uuid.Nil, err
	}

	if customer.LogoID != uuid.Nil {
		err = ci.driver.FileReference().AddToUsedBy(ctx, customer.LogoID, customer.ID)
		if err != nil {
			return uuid.Nil, err
		}
	}

	return customer.ID, nil
}

// Update modifies an existing customer in the database
func (ci *CustomerIndex) Update(ctx context.Context, customerID uuid.UUID, customer *Customer) error {
	filter := bson.M{"_id": customerID}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
			"name":       customer.Name,
			"language":   customer.Language,
		},
	}

	_, err := ci.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

// UpdateLogo modifies an existing customer's logo
//
// Requires transactional context to ensure data integrity
func (ci *CustomerIndex) UpdateLogo(ctx context.Context, customerID uuid.UUID, logoID uuid.UUID, mime string) error {
	oldCustomer, err := ci.GetByID(ctx, customerID)
	if err != nil {
		return err
	}

	if oldCustomer.LogoID != uuid.Nil && oldCustomer.LogoID != logoID {
		err = ci.driver.FileReference().PullUsedBy(ctx, oldCustomer.LogoID, oldCustomer.ID)
		if err != nil {
			return err
		}
	}

	filter := bson.M{"_id": customerID}

	update := bson.M{
		"$set": bson.M{
			"updated_at":     time.Now(),
			"logo_id":        logoID,
			"logo_mime_type": mime,
			"logo_reference": util.CreateImageReference(mime, logoID),
		},
	}

	_, err = ci.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if logoID != uuid.Nil {
		err = ci.driver.FileReference().AddToUsedBy(ctx, logoID, oldCustomer.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete removes a customer and all its associated data
//
// Requires transactional context to ensure data integrity
func (ci *CustomerIndex) Delete(ctx context.Context, customerID uuid.UUID) error {
	// retrieve the current customer document
	oldCustomer, err := ci.GetByID(ctx, customerID)
	if err != nil {
		return err
	}

	// delete logo reference
	if oldCustomer.LogoID != uuid.Nil {
		err = ci.driver.FileReference().PullUsedBy(ctx, oldCustomer.LogoID, oldCustomer.ID)
		if err != nil {
			return err
		}
	}

	// Remove the customer from the user's list
	filter := bson.M{"customers._id": customerID}
	update := bson.M{"$pull": bson.M{"customers": bson.M{"_id": customerID}}}
	_, err = ci.driver.User().collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	// Remove all targets for the customer
	targets, err := ci.driver.Target().GetByCustomerID(ctx, customerID)
	if err != nil {
		return err
	}

	for _, target := range targets {
		if err := ci.driver.Target().Delete(ctx, target.ID); err != nil {
			return err
		}
	}

	// Remove all templates for the customer
	templates, err := ci.driver.Template().GetByCustomerID(ctx, customerID)
	if err != nil {
		return err
	}

	for _, template := range templates {
		if err := ci.driver.Template().Delete(ctx, template.ID); err != nil {
			return err
		}
	}

	// Remove all assessments for the customer
	assessments, err := ci.driver.Assessment().GetByCustomerID(ctx, customerID)
	if err != nil {
		return err
	}

	for _, assessment := range assessments {
		if err := ci.driver.Assessment().Delete(ctx, assessment.ID); err != nil {
			return fmt.Errorf("failed to delete Assessment %s: %w", assessment.ID, err)
		}
	}

	// Delete the customer
	_, err = ci.collection.DeleteOne(ctx, bson.M{"_id": customerID})
	return err
}

func (ci *CustomerIndex) GetByID(ctx context.Context, customerID uuid.UUID) (*Customer, error) {
	var customer Customer
	if err := ci.collection.FindOne(ctx, bson.M{"_id": customerID}).Decode(&customer); err != nil {
		return nil, err
	}
	return &customer, nil
}

func (ci *CustomerIndex) GetByIDForHydrate(ctx context.Context, customerID uuid.UUID) (*Customer, error) {
	filter := bson.M{"_id": customerID}
	opts := options.FindOne().SetProjection(bson.M{
		"logo_id":   0,
		"templates": 0,
	})

	var customer Customer
	err := ci.collection.FindOne(ctx, filter, opts).Decode(&customer)
	if err != nil {
		return nil, err
	}

	return &customer, nil
}

func (ci *CustomerIndex) GetManyForHydrate(ctx context.Context, customers []Customer) ([]Customer, error) {
	customerIDs := make([]uuid.UUID, len(customers))
	for i := range customers {
		customerIDs[i] = customers[i].ID
	}

	filter := bson.M{"_id": bson.M{"$in": customerIDs}}
	opts := options.Find().SetProjection(bson.M{
		"logo_id":   0,
		"templates": 0,
	})

	cursor, err := ci.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	customersMongo := []Customer{}
	err = cursor.All(ctx, &customersMongo)
	if err != nil {
		return nil, err
	}

	return customersMongo, nil
}

func (ci *CustomerIndex) GetByIDPipeline(ctx context.Context, customerID uuid.UUID) (*Customer, error) {
	filter := bson.M{"_id": customerID}

	customer := &Customer{}
	err := ci.collection.FindOne(ctx, filter).Decode(customer)
	if err != nil {
		return nil, err
	}

	err = ci.hydrate(ctx, customer)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

func (ci *CustomerIndex) GetAll(ctx context.Context, customerIDs []uuid.UUID) ([]Customer, error) {
	filter := bson.M{}
	if customerIDs != nil {
		filter = bson.M{
			"_id": bson.M{"$in": customerIDs},
		}
	}

	cursor, err := ci.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	customers := []Customer{}
	err = cursor.All(ctx, &customers)
	if err != nil {
		return nil, err
	}

	for i := range customers {
		err = ci.hydrate(ctx, &customers[i])
		if err != nil {
			return nil, err
		}
	}

	return customers, nil
}

func (ci *CustomerIndex) Search(ctx context.Context, query string) ([]Customer, error) {
	filter := bson.M{
		"name": bson.M{"$regex": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}},
	}

	cursor, err := ci.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	customers := []Customer{}
	err = cursor.All(ctx, &customers)
	if err != nil {
		return nil, err
	}

	return customers, nil
}

// hydrate fills in the nested fields for a Customer
func (ci *CustomerIndex) hydrate(ctx context.Context, customer *Customer) error {
	templates, err := ci.driver.Template().GetByCustomerIDForHydrate(ctx, customer.ID)
	if err != nil {
		return err
	}

	customer.Templates = templates

	return nil
}
