package db

import (
	"context"
	"time"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AssessmentIndex struct{ driver *Driver }

func (ai *AssessmentIndex) selectWithCustomer(ctx context.Context, row any) *bun.SelectQuery {
	return idbFrom(ctx, ai.driver.db).NewSelect().Model(row).Relation("Customer")
}

func (ai *AssessmentIndex) selectWithRelations(ctx context.Context, row any) *bun.SelectQuery {
	return ai.selectWithCustomer(ctx, row).Relation("Targets")
}

func rowToAssessmentWithRelations(r *dbAssessment) model.Assessment {
	a := r.toModel()
	a.Targets = make([]model.Target, len(r.Targets))
	for i := range r.Targets {
		a.Targets[i] = r.Targets[i].toModel()
	}
	return a
}

func (ai *AssessmentIndex) Insert(ctx context.Context, assessment *model.Assessment, customerID uuid.UUID) (uuid.UUID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}
	row := &dbAssessment{
		ID:              id,
		CustomerID:      customerID,
		Name:            assessment.Name,
		Language:        assessment.Language,
		StartDateTime:   timePtrIfSet(assessment.StartDateTime),
		EndDateTime:     timePtrIfSet(assessment.EndDateTime),
		KickoffDateTime: timePtrIfSet(assessment.KickoffDateTime),
		Status:          assessment.Status,
		TypeShort:       assessment.Type.Short,
		TypeFull:        assessment.Type.Full,
		CVSSVersions:    boolMap(assessment.CVSSVersions),
		Environment:     assessment.Environment,
		TestingType:     assessment.TestingType,
		OSSTMMVector:    assessment.OSSTMMVector,
	}
	if _, err := idbFrom(ctx, ai.driver.db).NewInsert().Model(row).Exec(ctx); err != nil {
		return uuid.Nil, mapErr(err)
	}
	if err := ai.insertAssessmentTargets(ctx, id, assessment.Targets); err != nil {
		return uuid.Nil, err
	}
	assessment.ID = id
	assessment.Customer.ID = customerID
	return id, nil
}

func (ai *AssessmentIndex) insertAssessmentTargets(ctx context.Context, assessmentID uuid.UUID, targets []model.Target) error {
	rows := make([]dbAssessmentTarget, 0, len(targets))
	for _, t := range targets {
		if t.ID == uuid.Nil {
			continue
		}
		rows = append(rows, dbAssessmentTarget{AssessmentID: assessmentID, TargetID: t.ID})
	}
	if len(rows) == 0 {
		return nil
	}
	_, err := idbFrom(ctx, ai.driver.db).NewInsert().
		Model(&rows).
		On("CONFLICT DO NOTHING").
		Exec(ctx)
	return mapErr(err)
}

func (ai *AssessmentIndex) GetByID(ctx context.Context, id uuid.UUID) (*model.Assessment, error) {
	var row dbAssessment
	if err := ai.selectWithCustomer(ctx, &row).
		Where("a.id = ?", id).
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := row.toModel()
	return &out, nil
}

func (ai *AssessmentIndex) GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*model.Assessment, error) {
	var row dbAssessment
	if err := ai.selectWithRelations(ctx, &row).
		Where("a.id = ?", id).
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := rowToAssessmentWithRelations(&row)
	return &out, nil
}

func (ai *AssessmentIndex) GetMultipleByID(ctx context.Context, ids []uuid.UUID) ([]model.Assessment, error) {
	var rows []dbAssessment
	if err := ai.selectWithRelations(ctx, &rows).
		Where("a.id IN (?)", bun.List(ids)).
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := make([]model.Assessment, len(rows))
	for i := range rows {
		out[i] = rowToAssessmentWithRelations(&rows[i])
	}
	return out, nil
}

func (ai *AssessmentIndex) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]model.Assessment, error) {
	var rows []dbAssessment
	if err := ai.selectWithRelations(ctx, &rows).
		Where("a.customer_id = ?", customerID).
		OrderExpr("a.created_at DESC").
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := make([]model.Assessment, len(rows))
	for i := range rows {
		out[i] = rowToAssessmentWithRelations(&rows[i])
	}
	return out, nil
}

func (ai *AssessmentIndex) Search(ctx context.Context, customers []uuid.UUID, customerID uuid.UUID, name string) ([]model.Assessment, error) {
	var rows []dbAssessment
	q := ai.selectWithRelations(ctx, &rows)
	if name != "" {
		q = q.Where("a.name ILIKE ?", "%"+name+"%")
	}
	switch {
	case customerID != uuid.Nil:
		q = q.Where("a.customer_id = ?", customerID)
	case len(customers) > 0:
		q = q.Where("a.customer_id IN (?)", bun.List(customers))
	}
	q = q.OrderExpr("a.name")

	if err := q.Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	out := make([]model.Assessment, len(rows))
	for i := range rows {
		out[i] = rowToAssessmentWithRelations(&rows[i])
	}
	return out, nil
}

func (ai *AssessmentIndex) Update(ctx context.Context, id uuid.UUID, assessment *model.Assessment) error {
	idb := idbFrom(ctx, ai.driver.db)
	if _, err := idb.NewUpdate().
		Model((*dbAssessment)(nil)).
		Set("name = ?", assessment.Name).
		Set("language = ?", assessment.Language).
		Set("start_date_time = ?", timePtrIfSet(assessment.StartDateTime)).
		Set("end_date_time = ?", timePtrIfSet(assessment.EndDateTime)).
		Set("kickoff_date_time = ?", timePtrIfSet(assessment.KickoffDateTime)).
		Set("status = ?", assessment.Status).
		Set("type_short = ?", assessment.Type.Short).
		Set("type_full = ?", assessment.Type.Full).
		Set("cvss_versions = ?", boolMap(assessment.CVSSVersions)).
		Set("environment = ?", assessment.Environment).
		Set("testing_type = ?", assessment.TestingType).
		Set("osstmm_vector = ?", assessment.OSSTMMVector).
		Where("id = ?", id).
		Exec(ctx); err != nil {
		return mapErr(err)
	}

	if assessment.Targets != nil {
		if _, err := idb.NewDelete().
			Model((*dbAssessmentTarget)(nil)).
			Where("assessment_id = ?", id).
			Exec(ctx); err != nil {
			return mapErr(err)
		}
		if err := ai.insertAssessmentTargets(ctx, id, assessment.Targets); err != nil {
			return err
		}
	}
	return nil
}

func (ai *AssessmentIndex) UpdateStatus(ctx context.Context, id uuid.UUID, assessment *model.Assessment) error {
	_, err := idbFrom(ctx, ai.driver.db).NewUpdate().
		Model((*dbAssessment)(nil)).
		Set("status = ?", assessment.Status).
		Where("id = ?", id).
		Exec(ctx)
	return mapErr(err)
}

func (ai *AssessmentIndex) UpdateTargets(ctx context.Context, id uuid.UUID, targetID uuid.UUID) error {
	_, err := idbFrom(ctx, ai.driver.db).NewInsert().
		Model(&dbAssessmentTarget{AssessmentID: id, TargetID: targetID}).
		On("CONFLICT DO NOTHING").
		Exec(ctx)
	return mapErr(err)
}

func (ai *AssessmentIndex) BulkUpdateTargets(ctx context.Context, id uuid.UUID, targetIDs []uuid.UUID) error {
	if len(targetIDs) == 0 {
		return nil
	}
	rows := make([]dbAssessmentTarget, len(targetIDs))
	for i, tid := range targetIDs {
		rows[i] = dbAssessmentTarget{AssessmentID: id, TargetID: tid}
	}
	_, err := idbFrom(ctx, ai.driver.db).NewInsert().
		Model(&rows).
		On("CONFLICT DO NOTHING").
		Exec(ctx)
	return mapErr(err)
}

func (ai *AssessmentIndex) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := idbFrom(ctx, ai.driver.db).NewDelete().
		Model((*dbAssessment)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return mapErr(err)
}

func (ai *AssessmentIndex) Clone(ctx context.Context, sourceID uuid.UUID, name string, includePocs bool) (uuid.UUID, error) {
	newID, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}
	idb := idbFrom(ctx, ai.driver.db)

	if _, err := idb.NewRaw(`
		INSERT INTO assessment (id, customer_id, name, language, start_date_time, end_date_time,
			kickoff_date_time, status, type_short, type_full, cvss_versions, environment,
			testing_type, osstmm_vector, vulnerability_count)
		SELECT ?, customer_id, ?, language, start_date_time, end_date_time,
			kickoff_date_time, status, type_short, type_full, cvss_versions, environment,
			testing_type, osstmm_vector, 0
		FROM assessment WHERE id = ?
	`, newID, name, sourceID).Exec(ctx); err != nil {
		return uuid.Nil, mapErr(err)
	}

	if _, err := idb.NewRaw(`
		INSERT INTO assessment_target (assessment_id, target_id)
		SELECT ?, target_id FROM assessment_target WHERE assessment_id = ?
	`, newID, sourceID).Exec(ctx); err != nil {
		return uuid.Nil, mapErr(err)
	}

	vulns, err := ai.driver.Vulnerability().GetByAssessmentID(ctx, sourceID)
	if err != nil {
		return uuid.Nil, err
	}
	for _, v := range vulns {
		if _, err := ai.driver.Vulnerability().Clone(ctx, v.ID, newID, includePocs); err != nil {
			return uuid.Nil, err
		}
	}
	return newID, nil
}

func timePtrIfSet(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func boolMap(m map[string]bool) map[string]bool {
	if m == nil {
		return map[string]bool{}
	}
	return m
}
