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
	targetCollection = "target"
)

type Target struct {
	Model    `bson:",inline"`
	IPv4     string   `json:"ipv4,omitempty" bson:"ipv4"`
	IPv6     string   `json:"ipv6,omitempty" bson:"ipv6"`
	Port     int      `json:"port,omitempty" bson:"port"`
	Protocol string   `json:"protocol,omitempty" bson:"protocol"`
	FQDN     string   `json:"fqdn" bson:"fqdn"`
	Tag      string   `json:"tag,omitempty" bson:"tag"`
	Customer Customer `json:"customer,omitempty" bson:"customer"`
}

type TargetIndex struct {
	driver     *Driver
	collection *mongo.Collection
}

func (d *Driver) Target() *TargetIndex {
	return &TargetIndex{
		driver:     d,
		collection: d.database.Collection(targetCollection),
	}
}

func (ti TargetIndex) init() error {
	_, err := ti.collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "ipv4", Value: 1},
				{Key: "ipv6", Value: 1},
				{Key: "fqdn", Value: 1},
				{Key: "tag", Value: 1},
				{Key: "customer._id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}

func (ti *TargetIndex) Insert(ctx context.Context, target *Target, customerID uuid.UUID) (uuid.UUID, error) {
	err := ti.driver.Customer().collection.FindOne(ctx, bson.M{"_id": customerID}).Err()
	if err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	target.Model = Model{
		ID:        id,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	target.Customer = Customer{
		Model: Model{
			ID: customerID,
		},
	}

	_, err = ti.collection.InsertOne(ctx, target)
	if err != nil {
		return uuid.Nil, err
	}

	return target.ID, nil
}

func (ti *TargetIndex) FirstOrInsert(ctx context.Context, target *Target, customerID uuid.UUID) (uuid.UUID, bool, error) {
	err := ti.driver.Customer().collection.FindOne(ctx, bson.M{"_id": customerID}).Err()
	if err != nil {
		return uuid.Nil, false, err
	}

	var existingTarget Assessment
	err = ti.collection.FindOne(ctx, bson.M{
		"ipv4":         target.IPv4,
		"ipv6":         target.IPv6,
		"fqdn":         target.FQDN,
		"tag":          target.Tag,
		"customer._id": customerID,
	}).Decode(&existingTarget)
	if err == nil {
		return existingTarget.ID, false, nil
	}
	if err != mongo.ErrNoDocuments {
		return uuid.Nil, false, err
	}

	id, err := ti.Insert(ctx, target, customerID)
	return id, true, err
}

func (ti *TargetIndex) Update(ctx context.Context, targetID uuid.UUID, target *Target) error {
	if targetID == ImmutableID {
		return ErrImmutableTarget
	}

	filter := bson.M{"_id": targetID}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
			"ipv4":       target.IPv4,
			"ipv6":       target.IPv6,
			"port":       target.Port,
			"protocol":   target.Protocol,
			"fqdn":       target.FQDN,
			"tag":        target.Tag,
		},
	}

	_, err := ti.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete removes a target. It also removes references to this
// target from vulnerabilities and assessments.
//
// Requires transactional context to ensure data integrity
func (ti *TargetIndex) Delete(ctx context.Context, targetID uuid.UUID) error {
	if targetID == ImmutableID {
		return ErrImmutableTarget
	}

	filter := bson.M{"target._id": targetID}
	update := bson.M{
		"$set": bson.M{
			"target._id": ImmutableID,
		},
	}
	_, err := ti.driver.Vulnerability().collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	filter = bson.M{"targets._id": targetID}
	update = bson.M{
		"$pull": bson.M{
			"targets": bson.M{"_id": targetID},
		},
	}
	_, err = ti.driver.Assessment().collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	_, err = ti.collection.DeleteOne(ctx, bson.M{"_id": targetID})
	return err
}

func (ti *TargetIndex) GetByID(ctx context.Context, targetID uuid.UUID) (*Target, error) {
	var target Target
	err := ti.collection.FindOne(ctx, bson.M{"_id": targetID}).Decode(&target)
	if err != nil {
		return nil, err
	}

	return &target, nil
}
func (ti *TargetIndex) GetByIDPipeline(ctx context.Context, targetID uuid.UUID) (*Target, error) {
	filter := bson.M{"_id": targetID}

	target := &Target{}
	err := ti.collection.FindOne(ctx, filter).Decode(target)
	if err != nil {
		return nil, err
	}

	err = ti.hydrate(ctx, target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

func (ti *TargetIndex) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]Target, error) {
	filter := bson.M{
		"$and": []bson.M{
			{"customer._id": customerID},
			{"_id": bson.M{"$ne": ImmutableID}},
		},
	}

	cursor, err := ti.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	targets := []Target{}
	err = cursor.All(ctx, &targets)
	if err != nil {
		return nil, err
	}

	for i := range targets {
		err = ti.hydrate(ctx, &targets[i])
		if err != nil {
			return nil, err
		}
	}

	return targets, nil
}

func (ti *TargetIndex) Search(ctx context.Context, customerID uuid.UUID, query string) ([]Target, error) {
	conditions := []bson.M{
		{"_id": bson.M{"$ne": ImmutableID}},
	}

	if customerID != uuid.Nil {
		conditions = append(conditions, bson.M{"customer._id": customerID})
	}

	orCondition := bson.M{}
	if query != "" {
		orCondition = bson.M{
			"$or": []bson.M{
				{"ipv4": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}},
				{"ipv6": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}},
				{"port": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}},
				{"protocol": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}},
				{"fqdn": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}},
				{"tag": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}},
			},
		}
	}
	conditions = append(conditions, orCondition)

	filter := bson.M{
		"$and": conditions,
	}

	cursor, err := ti.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	targets := []Target{}
	err = cursor.All(ctx, &targets)
	if err != nil {
		return nil, err
	}

	for i := range targets {
		err = ti.hydrate(ctx, &targets[i])
		if err != nil {
			return nil, err
		}
	}

	return targets, nil
}

func (ti *TargetIndex) SearchWithinCustomers(ctx context.Context, customerIDs []uuid.UUID, query string) ([]Target, error) {
	conditions := []bson.M{
		{"_id": bson.M{"$ne": ImmutableID}},
	}

	if customerIDs != nil {
		conditions = append(conditions, bson.M{"customer._id": bson.M{"$in": customerIDs}})
	}

	orCondition := bson.M{}
	if query != "" {
		orCondition = bson.M{
			"$or": []bson.M{
				{"ipv4": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}},
				{"ipv6": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}},
				{"port": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}},
				{"protocol": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}},
				{"fqdn": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}},
				{"tag": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}},
			},
		}
	}
	conditions = append(conditions, orCondition)

	filter := bson.M{
		"$and": conditions,
	}

	cursor, err := ti.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	targets := []Target{}
	err = cursor.All(ctx, &targets)
	if err != nil {
		return nil, err
	}

	for i := range targets {
		err = ti.hydrate(ctx, &targets[i])
		if err != nil {
			return nil, err
		}
	}

	return targets, nil
}

// hydrate fills in the nested fields for a Target
func (ti *TargetIndex) hydrate(ctx context.Context, target *Target) error {
	customer, err := ti.driver.Customer().GetByIDForHydrate(ctx, target.Customer.ID)
	if err != nil {
		customer = &Customer{}
	}

	target.Customer = *customer

	return nil
}
