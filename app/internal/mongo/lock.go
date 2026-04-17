package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	lockCollection = "__lock"

	LockAdmin    = "admin-lock"
	LockUsername = "username-lock"
)

type Lock struct {
	Model       `bson:",inline"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	LockedBy    uuid.UUID `bson:"locked_by"`
	LockedAt    time.Time `bson:"locked_at"`
}

type LockIndex struct {
	driver     *Driver
	collection *mongo.Collection
}

func (d *Driver) Lock() *LockIndex {
	return &LockIndex{
		driver:     d,
		collection: d.database.Collection(lockCollection),
	}
}

func (ui LockIndex) init() error {
	_, err := ui.collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "name", Value: 1},
			},
		},
	)
	return err
}

func (ui *LockIndex) Lock(ctx context.Context, lockName string, lockedBy uuid.UUID) error {
	if lockName == "" {
		return nil
	}

	ui.driver.logger.Info().Msgf("Attempting to acquire lock \"%s\"", lockName)

	now := time.Now()
	staleThreshold := now.Add(-1 * time.Minute)

	filter := bson.M{
		"$or": []bson.M{
			{"name": lockName, "locked_at": bson.M{"$gt": staleThreshold}},
			{"name": lockName, "locked_at": bson.M{"$exists": false}},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"name":       lockName,
			"locked_by":  lockedBy,
			"locked_at":  now,
			"updated_at": now,
		},
		"$setOnInsert": bson.M{
			"created_at": now,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)

	res, err := ui.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to update lock: %w", err)
	}

	if res.MatchedCount > 0 {
		return fmt.Errorf("%w: lock \"%s\" is already held", ErrLocked, lockName)
	}

	ui.driver.logger.Info().Msgf("Lock \"%s\" successfully acquired by %s", lockName, lockedBy)
	return nil
}

func (ui *LockIndex) Unlock(ctx context.Context, lockName string) error {
	if lockName == "" {
		return nil
	}

	filter := bson.M{
		"name": lockName,
	}

	res, err := ui.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
