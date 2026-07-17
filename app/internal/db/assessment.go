package db

import (
	"context"

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
		CVSSVersions:    emptyMapIfNil(assessment.CVSSVersions),
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
		Set("cvss_versions = ?", emptyMapIfNil(assessment.CVSSVersions)).
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

	if includePocs {
		_, err = idb.NewRaw(cloneVulnerabilitiesWithPocsSQL, sourceID, newID).Exec(ctx)
	} else {
		_, err = idb.NewRaw(`
			INSERT INTO vulnerability (id, assessment_id, `+vulnCloneColumns+`)
			SELECT gen_random_uuid(), ?, `+vulnCloneColumns+`
			FROM vulnerability WHERE assessment_id = ?
		`, newID, sourceID).Exec(ctx)
	}
	if err != nil {
		return uuid.Nil, mapErr(err)
	}
	return newID, nil
}

// cloneVulnerabilitiesWithPocsSQL clones every vulnerability of an assessment
// together with its poc and poc_image rows in a single set-based statement.
const cloneVulnerabilitiesWithPocsSQL = `
	WITH vuln_map AS MATERIALIZED (
		SELECT id AS old_id, gen_random_uuid() AS new_id
		FROM vulnerability
		WHERE assessment_id = ?
	),
	ins_vuln AS (
		INSERT INTO vulnerability (id, assessment_id, ` + vulnCloneColumns + `)
		SELECT m.new_id, ?, ` + vulnCloneColumns + `
		FROM vulnerability
		JOIN vuln_map m ON m.old_id = vulnerability.id
	),
	poc_map AS MATERIALIZED (
		SELECT gen_random_uuid() AS new_poc_id, m.new_id AS new_vulnerability_id, p.items
		FROM poc p
		JOIN vuln_map m ON m.old_id = p.vulnerability_id
	),
	item_map AS MATERIALIZED (
		SELECT pm.new_poc_id, t.item, t.ord, gen_random_uuid() AS new_item_id
		FROM poc_map pm
		CROSS JOIN LATERAL jsonb_array_elements(pm.items) WITH ORDINALITY AS t(item, ord)
	),
	ins_poc AS (
		INSERT INTO poc (id, vulnerability_id, items)
		SELECT pm.new_poc_id, pm.new_vulnerability_id,
			COALESCE((
				SELECT jsonb_agg(jsonb_set(im.item, '{id}', to_jsonb(im.new_item_id)) ORDER BY im.ord)
				FROM item_map im
				WHERE im.new_poc_id = pm.new_poc_id
			), '[]'::jsonb)
		FROM poc_map pm
	)
	INSERT INTO poc_image (poc_id, poc_item_id, file_reference_id)
	SELECT im.new_poc_id, im.new_item_id, (im.item->>'image_id')::uuid
	FROM item_map im
	WHERE COALESCE(im.item->>'image_id', '')
		NOT IN ('', '00000000-0000-0000-0000-000000000000')
`
