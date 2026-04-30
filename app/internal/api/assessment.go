package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Kryvea/Kryvea/internal/cvss"
	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/report"
	reportdata "github.com/Kryvea/Kryvea/internal/report/data"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type assessmentRequestData struct {
	Name            string               `json:"name"`
	Language        string               `json:"language"`
	StartDateTime   time.Time            `json:"start_date_time"`
	EndDateTime     time.Time            `json:"end_date_time"`
	KickoffDateTime time.Time            `json:"kickoff_date_time"`
	Status          string               `json:"status"`
	Targets         []string             `json:"targets"`
	Type            model.AssessmentType `json:"type"`
	CVSSVersions    map[string]bool      `json:"cvss_versions"`
	Environment     string               `json:"environment"`
	TestingType     string               `json:"testing_type"`
	OSSTMMVector    string               `json:"osstmm_vector"`
	CustomerID      string               `json:"customer_id"`
}

func (d *Driver) AddAssessment(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// parse request body
	data := &assessmentRequestData{}
	if err := c.BodyParser(data); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// check if user has access to customer
	customer, errStr := d.customerFromParam(c.UserContext(), data.CustomerID)
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	if !user.CanAccessCustomer(customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	// validate data
	errStr = d.validateAssessmentData(data)
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// parse targets
	targets := []model.Target{}
	for _, target := range data.Targets {
		targetID, err := util.ParseUUID(target)
		if err != nil {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": "Invalid target ID",
			})
		}
		targets = append(targets, model.Target{
			Model: model.Model{
				ID: targetID,
			},
		})
	}

	assessment := &model.Assessment{
		Name:            data.Name,
		Language:        data.Language,
		StartDateTime:   data.StartDateTime,
		EndDateTime:     data.EndDateTime,
		KickoffDateTime: data.KickoffDateTime,
		Targets:         targets,
		Status:          data.Status,
		Type:            data.Type,
		CVSSVersions:    data.CVSSVersions,
		Environment:     data.Environment,
		TestingType:     data.TestingType,
		OSSTMMVector:    data.OSSTMMVector,
	}

	// insert assessment into database
	assessmentID, err := d.db.Assessment().Insert(c.UserContext(), assessment, customer.ID)
	if err != nil {
		c.Status(fiber.StatusBadRequest)

		if errors.Is(err, store.ErrDuplicateKey) {
			return c.JSON(fiber.Map{
				"error": fmt.Sprintf("Assessment \"%s\" already exists under customer \"%s\"", assessment.Name, customer.Name),
			})
		}

		return c.JSON(fiber.Map{
			"error": "Cannot create assessment",
		})
	}

	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"message":       "Assessment created",
		"assessment_id": assessmentID,
	})
}

func (d *Driver) SearchAssessments(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// parse query parameters
	customerParam := c.Query("customer")
	nameParam := c.Query("name")

	var customerID uuid.UUID
	if customerParam != "" {
		// check if user can access the customer
		customer, errStr := d.customerFromParam(c.UserContext(), customerParam)
		if errStr != "" {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": errStr,
			})
		}

		if !user.CanAccessCustomer(customer.ID) {
			c.Status(fiber.StatusForbidden)
			return c.JSON(fiber.Map{
				"error": "Forbidden",
			})
		}

		customerID = customer.ID
	}

	// retrieve user's customers
	customers := []uuid.UUID{}
	for _, uc := range user.Customers {
		customers = append(customers, uc.ID)
	}
	if user.Role == model.RoleAdmin {
		customers = nil
	}

	// retrieve assessments
	assessments, err := d.db.Assessment().Search(c.UserContext(), customers, customerID, nameParam)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot search assessments",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(assessments)
}

func (d *Driver) GetAssessmentsByCustomer(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// check if user can access the customer
	customer, errStr := d.customerFromParam(c.UserContext(), c.Params("customer"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	if !user.CanAccessCustomer(customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	// retrieve assessments
	assessments, err := d.db.Assessment().GetByCustomerID(c.UserContext(), customer.ID)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot retrieve assessments",
		})
	}

	// set owned assessments
	owned := make(map[uuid.UUID]struct{}, len(user.Assessments))
	for _, ua := range user.Assessments {
		owned[ua.ID] = struct{}{}
	}
	for i := range assessments {
		if _, ok := owned[assessments[i].ID]; ok {
			assessments[i].IsOwned = true
		}
	}

	c.Status(fiber.StatusOK)
	return c.JSON(assessments)
}

func (d *Driver) GetAssessment(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)
	// parse assessment param
	assessmentParam := c.Params("assessment")
	if assessmentParam == "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Assessment ID is required",
		})
	}

	assessmentID, err := util.ParseUUID(assessmentParam)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Invalid assessment ID",
		})
	}

	assessment, err := d.db.Assessment().GetByIDWithRelations(c.UserContext(), assessmentID)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Invalid assessment ID",
		})
	}

	// check if user has access to customer
	if !user.CanAccessCustomer(assessment.Customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	// set owned assessment
	for _, userAssessment := range user.Assessments {
		if userAssessment.ID == assessment.ID {
			assessment.IsOwned = true
			break
		}
	}

	c.Status(fiber.StatusOK)
	return c.JSON(assessment)
}

func (d *Driver) GetOwnedAssessments(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	if len(user.Assessments) == 0 {
		c.Status(fiber.StatusOK)
		return c.JSON([]model.Assessment{})
	}

	// map user assessments IDs
	userAssessments := make([]uuid.UUID, len(user.Assessments))
	for i, userAssessment := range user.Assessments {
		userAssessments[i] = userAssessment.ID
	}

	// get assessments from database
	assessments, err := d.db.Assessment().GetMultipleByID(c.UserContext(), userAssessments)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot get owned assessments",
		})
	}

	// set owned assessments
	owned := make(map[uuid.UUID]struct{}, len(user.Assessments))
	for _, ua := range user.Assessments {
		owned[ua.ID] = struct{}{}
	}
	for i := range assessments {
		if _, ok := owned[assessments[i].ID]; ok {
			assessments[i].IsOwned = true
		}
	}

	c.Status(fiber.StatusOK)
	return c.JSON(assessments)
}

func (d *Driver) UpdateAssessment(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// parse assessment param
	assessment, errStr := d.assessmentFromParam(c.UserContext(), c.Params("assessment"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// parse request body
	data := &assessmentRequestData{}
	if err := c.BodyParser(data); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// validate data
	errStr = d.validateAssessmentData(data)
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// check if user has access to customer
	if !user.CanAccessCustomer(assessment.Customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	// parse targets
	targets := []model.Target{}
	for _, target := range data.Targets {
		targetID, err := util.ParseUUID(target)
		if err != nil {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": "Invalid target ID",
			})
		}
		targets = append(targets, model.Target{
			Model: model.Model{
				ID: targetID,
			},
		})
	}

	newAssessment := &model.Assessment{
		Name:            data.Name,
		Language:        data.Language,
		StartDateTime:   data.StartDateTime,
		EndDateTime:     data.EndDateTime,
		KickoffDateTime: data.KickoffDateTime,
		Targets:         targets,
		Status:          data.Status,
		Type:            data.Type,
		CVSSVersions:    data.CVSSVersions,
		Environment:     data.Environment,
		TestingType:     data.TestingType,
		OSSTMMVector:    data.OSSTMMVector,
	}

	// update assessment in database
	err := d.db.Assessment().Update(c.UserContext(), assessment.ID, newAssessment)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)

		if errors.Is(err, store.ErrDuplicateKey) {
			return c.JSON(fiber.Map{
				"error": fmt.Sprintf("Assessment \"%s\" already exists under customer \"%s\"", newAssessment.Name, assessment.Customer.Name),
			})
		}

		return c.JSON(fiber.Map{
			"error": "Cannot update assessment",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Assessment updated",
	})
}

func (d *Driver) UpdateAssessmentStatus(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// parse assessment param
	assessment, errStr := d.assessmentFromParam(c.UserContext(), c.Params("assessment"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// parse request body
	data := &assessmentRequestData{}
	if err := c.BodyParser(data); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// validate data
	statusError := d.validateAssessmentStatus(data)
	if statusError != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Invalid status",
		})
	}

	// check if user has access to customer
	if !user.CanAccessCustomer(assessment.Customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	newAssessmentStatus := &model.Assessment{
		Status: data.Status,
	}

	// update assessment in database
	err := d.db.Assessment().UpdateStatus(c.UserContext(), assessment.ID, newAssessmentStatus)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot update assessment",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Assessment status updated",
	})
}

func (d *Driver) DeleteAssessment(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// parse assessment param
	assessment, errStr := d.assessmentFromParam(c.UserContext(), c.Params("assessment"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// check if user has access to customer
	if !user.CanAccessCustomer(assessment.Customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	if err := d.db.Assessment().Delete(c.UserContext(), assessment.ID); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot delete assessment",
		})
	}

	d.gcFilesAsync()
	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Assessment deleted",
	})
}

func (d *Driver) CloneAssessment(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// parse assessment param
	assessment, errStr := d.assessmentFromParam(c.UserContext(), c.Params("assessment"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// check if user has access to customer
	if !user.CanAccessCustomer(assessment.Customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	// parse request body
	type reqData struct {
		Name        string `json:"name"`
		IncludePocs bool   `json:"include_pocs"`
	}

	data := &reqData{}
	if err := c.BodyParser(data); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// validate data
	if data.Name == "" {
		data.Name = assessment.Name + " (Clone)"
	}

	cloneAssessmentID, err := d.db.RunInTx(c.UserContext(), func(ctx context.Context) (any, error) {
		// clone assessment
		cloneAssessmentID, err := d.db.Assessment().Clone(ctx, assessment.ID, data.Name, data.IncludePocs)
		if err != nil {
			if errors.Is(err, store.ErrDuplicateKey) {
				return uuid.Nil, fmt.Errorf("Assessment \"%s\" already exists", assessment.Name)
			}

			return uuid.Nil, errors.New("Cannot clone assessment")
		}

		return cloneAssessmentID, nil
	})
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message":       "Assessment cloned",
		"assessment_id": cloneAssessmentID.(uuid.UUID),
	})
}

type exportRequestData struct {
	Type                                string    `json:"type"`
	Template                            string    `json:"template"`
	DeliveryDateTime                    time.Time `json:"delivery_date_time"`
	SortByCvss                          string    `json:"sort_by_cvss"`
	IncludeInformationalVulnerabilities bool      `json:"include_informational_vulnerabilities"`
	FormatJson                          bool      `json:"format_json"`
}

func (d *Driver) ExportAssessment(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// parse assessment param
	assessmentParam := c.Params("assessment")
	if assessmentParam == "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Assessment ID is required",
		})
	}

	assessmentID, err := util.ParseUUID(assessmentParam)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Invalid Assessment ID",
		})
	}

	assessment, err := d.db.Assessment().GetByIDWithRelations(c.UserContext(), assessmentID)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Invalid Assessment ID",
		})
	}

	// check if user has access to customer
	customer, err := d.db.Customer().GetByID(c.UserContext(), assessment.Customer.ID)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Invalid customer ID",
		})
	}

	if !user.CanAccessCustomer(customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	if customer.LogoID != uuid.Nil {
		logoData, _, err := d.db.FileReference().ReadByID(c.UserContext(), customer.LogoID)
		if err != nil {
			c.Status(fiber.StatusInternalServerError)
			return c.JSON(fiber.Map{
				"error": "Failed to read logo data",
			})
		}
		customer.LogoData = logoData
	}

	// parse request body
	data := &exportRequestData{}
	if err := c.BodyParser(data); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// validate data
	if data.Type == "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Type is required",
		})
	}

	if !cvss.IsValidVersion(data.SortByCvss) {
		data.SortByCvss = util.GetMaxCvssVersion(assessment.CVSSVersions)
	}

	var templateBytes []byte

	if _, ok := report.ReportTemplateMap[data.Type]; ok {
		// validate template
		template, errStr := d.templateFromParam(c.UserContext(), data.Template)
		if errStr != "" {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": errStr,
			})
		}

		// retrieve template from database
		templateBytes, _, err = d.db.FileReference().ReadByID(c.UserContext(), template.FileID)
		if err != nil {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": "Invalid template ID",
			})
		}
	}

	// retrieve vulnerabilities
	vulnerabilities, err := d.db.Vulnerability().GetByAssessmentIDWithPocs(c.UserContext(), assessment.ID, data.IncludeInformationalVulnerabilities)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Failed to retrieve vulnerabilities",
		})
	}
	for i := range vulnerabilities {
		vulnerabilities[i].Assessment.Language = assessment.Language
	}
	model.HydrateCVSSAll(vulnerabilities)

	if err := d.loadPocImagesParallel(c.UserContext(), vulnerabilities); err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Failed to read image data",
		})
	}

	reportData := &reportdata.ReportData{
		Customer:         customer,
		Assessment:       assessment,
		Vulnerabilities:  vulnerabilities,
		DeliveryDateTime: data.DeliveryDateTime,
	}

	options := &reportdata.Options{
		FormatJson: data.FormatJson,
		SortByCvss: data.SortByCvss,
	}

	report, err := report.New(data.Type, templateBytes)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Invalid template type",
		})
	}

	// render report
	renderedTemplate, err := report.Render(reportData, options)
	if err != nil {
		d.logger.Error().Err(err).Msg("failed to render report")
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Failed to generate report",
			// TODO: remove
			"err": err.Error(),
		})
	}

	filename := util.SanitizeFileName(report.Filename())

	c.Status(fiber.StatusOK)
	c.Set("Content-Type", mimetype.Detect(renderedTemplate).String())
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	return c.SendStream(bytes.NewBuffer(renderedTemplate))
}

func (d *Driver) assessmentFromParam(ctx context.Context, assessmentParam string) (*model.Assessment, string) {
	if assessmentParam == "" {
		return nil, "Assessment ID is required"
	}

	assessmentID, err := util.ParseUUID(assessmentParam)
	if err != nil {
		return nil, "Invalid assessment ID"
	}

	assessment, err := d.db.Assessment().GetByID(ctx, assessmentID)
	if err != nil {
		return nil, "Invalid assessment ID"
	}

	return assessment, ""
}

func (d *Driver) validateAssessmentStatus(data *assessmentRequestData) string {
	switch data.Status {
	case
		model.ASSESSMENT_STATUS_ON_HOLD,
		model.ASSESSMENT_STATUS_IN_PROGRESS,
		model.ASSESSMENT_STATUS_COMPLETED:
	default:
		return "Invalid status"
	}

	return ""
}

func (d *Driver) validateAssessmentData(data *assessmentRequestData) string {
	if data.Name == "" {
		return "Name is required"
	}

	if data.Language == "" {
		return "Language is required"
	}

	if data.StartDateTime.IsZero() {
		return "Start date is required"
	}

	if data.EndDateTime.IsZero() {
		return "End date is required"
	}

	statusError := d.validateAssessmentStatus(data)
	if statusError != "" {
		return statusError
	}

	// filter valid cvssVersions
	data.CVSSVersions = d.filterValidCvssVersions(data.CVSSVersions)

	return ""
}

func (d *Driver) filterValidCvssVersions(cvssVersions map[string]bool) map[string]bool {
	validCvssVersions := make(map[string]bool)
	for _, version := range cvss.CvssVersions {
		validCvssVersions[version] = false
	}

	for version, enabled := range cvssVersions {
		if !enabled {
			continue
		}
		validCvssVersions[version] = true
	}

	return validCvssVersions
}

func (d *Driver) loadPocImagesParallel(parent context.Context, vulnerabilities []model.Vulnerability) error {
	const maxWorkers = 8

	type ref struct{ vi, pi int }
	var refs []ref
	for i := range vulnerabilities {
		for j, item := range vulnerabilities[i].Poc.Pocs {
			if item.ImageID != uuid.Nil {
				refs = append(refs, ref{vi: i, pi: j})
			}
		}
	}
	if len(refs) == 0 {
		return nil
	}

	sem := make(chan struct{}, maxWorkers)
	errCh := make(chan error, len(refs))
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	for _, r := range refs {
		wg.Add(1)
		sem <- struct{}{}
		go func(r ref) {
			defer wg.Done()
			defer func() { <-sem }()
			if ctx.Err() != nil {
				return
			}
			imageID := vulnerabilities[r.vi].Poc.Pocs[r.pi].ImageID
			data, _, err := d.db.FileReference().ReadByID(ctx, imageID)
			if err != nil {
				errCh <- err
				cancel()
				return
			}
			vulnerabilities[r.vi].Poc.Pocs[r.pi].ImageData = data
		}(r)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}
