package mongo

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	assessmentCollection = "assessment"
)

type Assessment struct {
	Model              `bson:",inline"`
	Name               string          `json:"name,omitempty" bson:"name"`
	Language           string          `json:"language,omitempty" bson:"language"`
	StartDateTime      time.Time       `json:"start_date_time,omitempty" bson:"start_date_time"`
	EndDateTime        time.Time       `json:"end_date_time,omitempty" bson:"end_date_time"`
	KickoffDateTime    time.Time       `json:"kickoff_date_time,omitempty" bson:"kickoff_date_time"`
	Targets            []Target        `json:"targets,omitempty" bson:"targets"`
	Status             string          `json:"status,omitempty" bson:"status"`
	Type               AssessmentType  `json:"type,omitempty" bson:"type"`
	CVSSVersions       map[string]bool `json:"cvss_versions,omitempty" bson:"cvss_versions"`
	Environment        string          `json:"environment,omitempty" bson:"environment"`
	TestingType        string          `json:"testing_type,omitempty" bson:"testing_type"`
	OSSTMMVector       string          `json:"osstmm_vector,omitempty" bson:"osstmm_vector"`
	VulnerabilityCount int             `json:"vulnerability_count,omitempty" bson:"vulnerability_count"`
	Customer           Customer        `json:"customer,omitempty" bson:"customer"`
	IsOwned            bool            `json:"is_owned,omitempty" bson:"is_owned"`
}

type AssessmentType struct {
	Short string `json:"short" bson:"short"`
	Full  string `json:"full" bson:"full"`
}

type AssessmentIndex struct {
	driver     *Driver
	collection *mongo.Collection
}

const (
	ASSESSMENT_STATUS_ON_HOLD     = "On Hold"
	ASSESSMENT_STATUS_IN_PROGRESS = "In Progress"
	ASSESSMENT_STATUS_COMPLETED   = "Completed"
)

func (d *Driver) Assessment() *AssessmentIndex {
	return &AssessmentIndex{
		driver:     d,
		collection: d.database.Collection(assessmentCollection),
	}
}

func (ai AssessmentIndex) init() error {
	_, err := ai.collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "name", Value: 1},
				{Key: "language", Value: 1},
				{Key: "customer._id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}

func (ai *AssessmentIndex) Insert(ctx context.Context, assessment *Assessment, customerID uuid.UUID) (uuid.UUID, error) {
	err := ai.driver.Customer().collection.FindOne(ctx, bson.M{"_id": customerID}).Err()
	if err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	assessment.Model = Model{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assessment.IsOwned = false
	assessment.Customer = Customer{
		Model: Model{
			ID: customerID,
		},
	}

	_, err = ai.collection.InsertOne(ctx, assessment)
	if err != nil {
		return uuid.Nil, err
	}

	return assessment.ID, nil
}

func (ai *AssessmentIndex) GetByID(ctx context.Context, assessmentID uuid.UUID) (*Assessment, error) {
	var assessment Assessment
	err := ai.collection.FindOne(ctx, bson.M{"_id": assessmentID}).Decode(&assessment)
	if err != nil {
		return nil, err
	}

	return &assessment, nil
}

func (ai *AssessmentIndex) GetByIDForHydrate(ctx context.Context, assessmentID uuid.UUID) (*Assessment, error) {
	filter := bson.M{"_id": assessmentID}
	opts := options.FindOne().SetProjection(bson.M{
		"name":          1,
		"language":      1,
		"cvss_versions": 1,
	})

	var assessment Assessment
	err := ai.collection.FindOne(ctx, filter, opts).Decode(&assessment)
	if err != nil {
		return nil, err
	}

	return &assessment, nil
}

func (ai *AssessmentIndex) GetManyForHydrate(ctx context.Context, assessments []Assessment) ([]Assessment, error) {
	assessmentIDs := make([]uuid.UUID, len(assessments))
	for i := range assessments {
		assessmentIDs[i] = assessments[i].ID
	}

	filter := bson.M{"_id": bson.M{"$in": assessmentIDs}}
	opts := options.Find().SetProjection(bson.M{
		"name": 1,
	})

	cursor, err := ai.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	assessmentsMongo := []Assessment{}
	err = cursor.All(ctx, &assessmentsMongo)
	if err != nil {
		return nil, err
	}

	return assessmentsMongo, nil
}

func (ai *AssessmentIndex) GetMultipleByID(ctx context.Context, assessmentIDs []uuid.UUID) ([]Assessment, error) {
	filter := bson.M{
		"_id": bson.M{"$in": assessmentIDs},
	}

	cursor, err := ai.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	assessment := []Assessment{}
	err = cursor.All(ctx, &assessment)
	if err != nil {
		return nil, err
	}

	for i := range assessment {
		err = ai.hydrate(ctx, &assessment[i])
		if err != nil {
			return nil, err
		}
	}

	return assessment, nil
}

func (ai *AssessmentIndex) GetByIDPipeline(ctx context.Context, assessmentID uuid.UUID) (*Assessment, error) {
	filter := bson.M{"_id": assessmentID}

	assessment := &Assessment{}
	err := ai.collection.FindOne(ctx, filter).Decode(assessment)
	if err != nil {
		return nil, err
	}

	err = ai.hydrate(ctx, assessment)
	if err != nil {
		return nil, err
	}

	return assessment, nil
}

func (ai *AssessmentIndex) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]Assessment, error) {
	filter := bson.M{"customer._id": customerID}

	cursor, err := ai.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	assessments := []Assessment{}
	err = cursor.All(ctx, &assessments)
	if err != nil {
		return nil, err
	}

	for i := range assessments {
		err = ai.hydrate(ctx, &assessments[i])
		if err != nil {
			return nil, err
		}
	}

	return assessments, nil
}

func (ai *AssessmentIndex) GetByCustomerAndID(ctx context.Context, customerID, assessmentID uuid.UUID) (*Assessment, error) {
	var assessment Assessment
	err := ai.collection.FindOne(ctx, bson.M{"_id": assessmentID, "customer._id": customerID}).Decode(&assessment)
	if err != nil {
		return nil, err
	}

	return &assessment, nil
}

func (ai *AssessmentIndex) Search(ctx context.Context, customers []uuid.UUID, customerID uuid.UUID, name string) ([]Assessment, error) {
	filter := bson.M{
		"name": bson.M{"$regex": bson.Regex{Pattern: regexp.QuoteMeta(name), Options: "i"}},
	}

	if customerID != uuid.Nil {
		filter["customer._id"] = customerID
	}

	if customerID == uuid.Nil && customers != nil {
		filter["customer._id"] = bson.M{"$in": customers}
	}

	cursor, err := ai.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	assessment := []Assessment{}
	err = cursor.All(ctx, &assessment)
	if err != nil {
		return nil, err
	}

	for i := range assessment {
		err = ai.hydrate(ctx, &assessment[i])
		if err != nil {
			return nil, err
		}
	}

	return assessment, nil
}

func (ai *AssessmentIndex) Update(ctx context.Context, assessmentID uuid.UUID, assessment *Assessment) error {
	filter := bson.M{"_id": assessmentID}

	update := bson.M{
		"$set": bson.M{
			"updated_at":        time.Now(),
			"name":              assessment.Name,
			"language":          assessment.Language,
			"start_date_time":   assessment.StartDateTime,
			"end_date_time":     assessment.EndDateTime,
			"kickoff_date_time": assessment.KickoffDateTime,
			"targets":           assessment.Targets,
			"status":            assessment.Status,
			"type":              assessment.Type,
			"cvss_versions":     assessment.CVSSVersions,
			"environment":       assessment.Environment,
			"testing_type":      assessment.TestingType,
			"osstmm_vector":     assessment.OSSTMMVector,
		},
	}

	_, err := ai.collection.UpdateOne(ctx, filter, update)
	return err
}

func (ai *AssessmentIndex) UpdateStatus(ctx context.Context, assessmentID uuid.UUID, assessment *Assessment) error {
	filter := bson.M{"_id": assessmentID}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
			"status":     assessment.Status,
		},
	}

	_, err := ai.collection.UpdateOne(ctx, filter, update)
	return err
}

func (ai *AssessmentIndex) UpdateTargets(ctx context.Context, assessmentID uuid.UUID, target uuid.UUID) error {
	filter := bson.M{"_id": assessmentID}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
		"$addToSet": bson.M{
			"targets": bson.M{
				"_id": target,
			},
		},
	}

	_, err := ai.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete removes an assessment and all its associated data
//
// Requires transactional context to ensure data integrity
func (ai *AssessmentIndex) Delete(ctx context.Context, assessmentID uuid.UUID) error {
	// TODO: move inside user index
	// Remove the assessment from the user's list
	filter := bson.M{"assessments._id": assessmentID}
	update := bson.M{"$pull": bson.M{"assessments": bson.M{"_id": assessmentID}}}
	_, err := ai.driver.User().collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to remove Assessment %s from Users: %w", assessmentID.String(), err)
	}

	// Delete all vulnerabilities associated with the assessment
	if err := ai.driver.Vulnerability().DeleteManyByAssessmentID(ctx, assessmentID); err != nil {
		return fmt.Errorf("failed to delete Vulnerabilities for Assessment %s: %w", assessmentID.String(), err)
	}

	// Delete the assessment
	_, err = ai.collection.DeleteOne(ctx, bson.M{"_id": assessmentID})
	return err
}

// Clone creates a copy of an assessment with the provided name
// including vulnerabilities and optionally PoCs
//
// Requires transactional context to ensure data integrity
func (ai *AssessmentIndex) Clone(ctx context.Context, assessmentID uuid.UUID, assessmentName string, includePocs bool) (uuid.UUID, error) {
	assessment, err := ai.GetByID(ctx, assessmentID)
	if err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	assessment.ID = id
	assessment.Name = assessmentName
	assessment.CreatedAt = time.Now()
	assessment.UpdatedAt = assessment.CreatedAt

	_, err = ai.collection.InsertOne(ctx, assessment)
	if err != nil {
		return uuid.Nil, err
	}

	// Clone vulnerabilities
	vulnerabilities, err := ai.driver.Vulnerability().GetByAssessmentID(ctx, assessmentID)
	if err != nil {
		return uuid.Nil, err
	}

	for _, vulnerability := range vulnerabilities {
		_, err := ai.driver.Vulnerability().Clone(ctx, vulnerability.ID, assessment.ID, includePocs)
		if err != nil {
			return uuid.Nil, err
		}
	}

	return assessment.ID, nil
}

// hydrate fills in the nested fields for an Assessment
func (ai *AssessmentIndex) hydrate(ctx context.Context, assessment *Assessment) error {
	customer, err := ai.driver.Customer().GetByIDForHydrate(ctx, assessment.Customer.ID)
	if err != nil {
		return err
	}

	assessment.Customer = *customer

	for i := range assessment.Targets {
		target, err := ai.driver.Target().GetByID(ctx, assessment.Targets[i].ID)
		if err != nil {
			return err
		}

		assessment.Targets[i] = *target
	}

	filter := bson.M{"assessment._id": assessment.ID}
	vulnCount, err := ai.driver.Vulnerability().collection.CountDocuments(ctx, filter)
	if err != nil {
		return err
	}

	assessment.VulnerabilityCount = int(vulnCount)

	return nil
}
