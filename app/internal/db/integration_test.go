//go:build integration

package db

import (
	"context"
	"os"
	"testing"

	"github.com/Kryvea/Kryvea/internal/log"
	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/google/uuid"
)

func newTestDriver(t *testing.T) *Driver {
	t.Helper()
	dsn := os.Getenv("KRYVEA_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("KRYVEA_TEST_PG_DSN not set; skipping integration tests")
	}

	tmp := t.TempDir()
	lw := log.NewLevelWriter(tmp, 1, 1, 1, false)

	d, err := NewDriver(context.Background(), dsn, tmp, "", "", lw)
	if err != nil {
		t.Fatalf("NewDriver: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return d
}

func uniqueName(prefix string) string {
	return prefix + "-" + uuid.NewString()
}

func seedAssessment(t *testing.T, d *Driver) (customerID, assessmentID, vulnID, target1ID, target2ID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	c := &model.Customer{Name: uniqueName("c"), Language: "en"}
	customerID, err := d.Customer().Insert(ctx, c)
	if err != nil {
		t.Fatalf("Customer.Insert: %v", err)
	}

	t1 := &model.Target{FQDN: uniqueName("t1") + ".example.com"}
	target1ID, err = d.Target().Insert(ctx, t1, customerID)
	if err != nil {
		t.Fatalf("Target.Insert#1: %v", err)
	}
	t2 := &model.Target{FQDN: uniqueName("t2") + ".example.com"}
	target2ID, err = d.Target().Insert(ctx, t2, customerID)
	if err != nil {
		t.Fatalf("Target.Insert#2: %v", err)
	}

	a := &model.Assessment{
		Name:     uniqueName("a"),
		Language: "en",
		Status:   model.ASSESSMENT_STATUS_IN_PROGRESS,
		Type:     model.AssessmentType{Short: "WAPT", Full: "Web App PT"},
		Targets:  []model.Target{{Model: model.Model{ID: target1ID}}, {Model: model.Model{ID: target2ID}}},
	}
	assessmentID, err = d.Assessment().Insert(ctx, a, customerID)
	if err != nil {
		t.Fatalf("Assessment.Insert: %v", err)
	}

	v := &model.Vulnerability{
		DetailedTitle: uniqueName("vuln"),
		Status:        model.VulnerabilityStatusOpen,
		Description:   "test description",
		Remediation:   "test remediation",
		References:    []string{"https://example.com"},
	}
	v.Assessment.ID = assessmentID
	v.Customer.ID = customerID
	v.Target.ID = target1ID
	v.Category.ID = model.ImmutableID
	vulnID, err = d.Vulnerability().Insert(ctx, v)
	if err != nil {
		t.Fatalf("Vulnerability.Insert: %v", err)
	}

	imageData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
		0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
		0x42, 0x60, 0x82,
	}
	imageID, mime, err := d.FileReference().Insert(ctx, imageData)
	if err != nil {
		t.Fatalf("FileReference.Insert: %v", err)
	}

	poc := &model.Poc{
		VulnerabilityID: vulnID,
		Pocs: []model.PocItem{
			{Index: 1, Type: "image", Description: "screenshot",
				ImageID: imageID, ImageMimeType: mime, ImageFilename: "shot.png"},
		},
	}
	if err := d.Poc().Upsert(ctx, poc); err != nil {
		t.Fatalf("Poc.Upsert: %v", err)
	}
	return
}

func TestIntegration_CloneAssessmentWithPoc(t *testing.T) {
	d := newTestDriver(t)
	ctx := context.Background()
	customerID, assessmentID, vulnID, target1ID, _ := seedAssessment(t, d)

	cloneName := uniqueName("clone")
	cloneID, err := d.Assessment().Clone(ctx, assessmentID, cloneName, true)
	if err != nil {
		t.Fatalf("Assessment.Clone: %v", err)
	}
	if cloneID == assessmentID {
		t.Fatal("clone returned the source ID")
	}

	cloned, err := d.Assessment().GetByIDWithRelations(ctx, cloneID)
	if err != nil {
		t.Fatalf("GetByIDWithRelations cloned: %v", err)
	}
	if cloned.Name != cloneName {
		t.Errorf("cloned name: got %q, want %q", cloned.Name, cloneName)
	}
	if cloned.Customer.ID != customerID {
		t.Errorf("cloned customer_id: got %s, want %s", cloned.Customer.ID, customerID)
	}
	if len(cloned.Targets) != 2 {
		t.Errorf("cloned targets: got %d, want 2", len(cloned.Targets))
	}

	cloneVulns, err := d.Vulnerability().GetByAssessmentID(ctx, cloneID)
	if err != nil {
		t.Fatalf("GetByAssessmentID cloned: %v", err)
	}
	if len(cloneVulns) != 1 {
		t.Fatalf("cloned vulns: got %d, want 1", len(cloneVulns))
	}
	if cloneVulns[0].ID == vulnID {
		t.Error("cloned vuln has same ID as source")
	}
	if cloneVulns[0].Target.ID != target1ID {
		t.Errorf("cloned vuln target_id: got %s, want %s", cloneVulns[0].Target.ID, target1ID)
	}

	clonePoc, err := d.Poc().GetByVulnerabilityID(ctx, cloneVulns[0].ID)
	if err != nil {
		t.Fatalf("GetByVulnerabilityID cloned poc: %v", err)
	}
	if len(clonePoc.Pocs) != 1 {
		t.Fatalf("cloned poc items: got %d, want 1", len(clonePoc.Pocs))
	}
	if clonePoc.Pocs[0].ImageID == uuid.Nil {
		t.Error("cloned poc lost the image_id reference")
	}
}

func TestIntegration_DeleteAssessmentCascade(t *testing.T) {
	d := newTestDriver(t)
	ctx := context.Background()
	_, assessmentID, vulnID, _, _ := seedAssessment(t, d)

	if err := d.Assessment().Delete(ctx, assessmentID); err != nil {
		t.Fatalf("Assessment.Delete: %v", err)
	}

	if _, err := d.Assessment().GetByID(ctx, assessmentID); err != store.ErrNotFound {
		t.Errorf("assessment after delete: got err=%v, want ErrNotFound", err)
	}
	if _, err := d.Vulnerability().GetByID(ctx, vulnID); err != store.ErrNotFound {
		t.Errorf("vulnerability after assessment delete: got err=%v, want ErrNotFound (cascade)", err)
	}

	pocs, err := d.Vulnerability().GetByAssessmentID(ctx, assessmentID)
	if err != nil {
		t.Fatalf("GetByAssessmentID after delete: %v", err)
	}
	if len(pocs) != 0 {
		t.Errorf("vulns after delete: got %d, want 0", len(pocs))
	}
}

func TestIntegration_DeleteCustomerCascade(t *testing.T) {
	d := newTestDriver(t)
	ctx := context.Background()
	customerID, assessmentID, vulnID, target1ID, target2ID := seedAssessment(t, d)

	poc, err := d.Poc().GetByVulnerabilityID(ctx, vulnID)
	if err != nil {
		t.Fatalf("Poc.GetByVulnerabilityID: %v", err)
	}
	imageID := poc.Pocs[0].ImageID

	if err := d.Customer().Delete(ctx, customerID); err != nil {
		t.Fatalf("Customer.Delete: %v", err)
	}

	if _, err := d.Customer().GetByID(ctx, customerID); err != store.ErrNotFound {
		t.Errorf("customer after delete: got err=%v, want ErrNotFound", err)
	}
	if _, err := d.Assessment().GetByID(ctx, assessmentID); err != store.ErrNotFound {
		t.Errorf("assessment after customer delete: got err=%v, want ErrNotFound", err)
	}
	if _, err := d.Vulnerability().GetByID(ctx, vulnID); err != store.ErrNotFound {
		t.Errorf("vulnerability after customer delete: got err=%v, want ErrNotFound", err)
	}
	for _, tid := range []uuid.UUID{target1ID, target2ID} {
		if _, err := d.Target().GetByIDWithRelations(ctx, tid); err != store.ErrNotFound {
			t.Errorf("target %s after customer delete: got err=%v, want ErrNotFound", tid, err)
		}
	}
	if _, err := d.FileReference().GetByID(ctx, imageID); err != store.ErrNotFound {
		t.Errorf("poc image file_reference after customer delete: got err=%v, want ErrNotFound", err)
	}
}

func TestIntegration_PocImageReplacement(t *testing.T) {
	d := newTestDriver(t)
	ctx := context.Background()
	_, _, vulnID, _, _ := seedAssessment(t, d)

	old, err := d.Poc().GetByVulnerabilityID(ctx, vulnID)
	if err != nil {
		t.Fatalf("Poc.GetByVulnerabilityID: %v", err)
	}
	oldImageID := old.Pocs[0].ImageID

	newPng := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x72, 0xB6, 0x0D,
		0x24, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9C, 0x63, 0x60, 0x00, 0x02, 0x00,
		0x00, 0x05, 0x00, 0x01, 0xE2, 0x26, 0x05, 0x9B,
		0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44,
		0xAE, 0x42, 0x60, 0x82,
	}
	newImageID, mime, err := d.FileReference().Insert(ctx, newPng)
	if err != nil {
		t.Fatalf("FileReference.Insert: %v", err)
	}
	if newImageID == oldImageID {
		t.Fatalf("expected distinct image ids")
	}

	updated := &model.Poc{
		VulnerabilityID: vulnID,
		Pocs: []model.PocItem{
			{Index: 1, Type: "image", Description: "screenshot",
				ImageID: newImageID, ImageMimeType: mime, ImageFilename: "shot2.png"},
		},
	}
	if err := d.Poc().Upsert(ctx, updated); err != nil {
		t.Fatalf("Poc.Upsert: %v", err)
	}

	if _, err := d.FileReference().GetByID(ctx, oldImageID); err != store.ErrNotFound {
		t.Errorf("old image file_reference after replacement: got err=%v, want ErrNotFound", err)
	}
	if _, err := d.FileReference().GetByID(ctx, newImageID); err != nil {
		t.Errorf("new image file_reference missing: %v", err)
	}
}

func TestIntegration_DeleteTargetReassign(t *testing.T) {
	d := newTestDriver(t)
	ctx := context.Background()
	_, assessmentID, vulnID, target1ID, _ := seedAssessment(t, d)

	if err := d.Target().Delete(ctx, target1ID); err != nil {
		t.Fatalf("Target.Delete: %v", err)
	}

	v, err := d.Vulnerability().GetByID(ctx, vulnID)
	if err != nil {
		t.Fatalf("GetByID after target delete: %v", err)
	}
	if v.Target.ID != model.ImmutableID {
		t.Errorf("vuln target_id after delete: got %s, want immutable %s", v.Target.ID, model.ImmutableID)
	}

	a, err := d.Assessment().GetByIDWithRelations(ctx, assessmentID)
	if err != nil {
		t.Fatalf("GetByIDWithRelations after target delete: %v", err)
	}
	for _, t2 := range a.Targets {
		if t2.ID == target1ID {
			t.Errorf("assessment_target still references deleted target %s", target1ID)
		}
	}
}
