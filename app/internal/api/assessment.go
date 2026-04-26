package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Kryvea/Kryvea/internal/cvss"
	"github.com/Kryvea/Kryvea/internal/mongo"
	"github.com/Kryvea/Kryvea/internal/report"
	reportdata "github.com/Kryvea/Kryvea/internal/report/data"
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
	Type            mongo.AssessmentType `json:"type"`
	CVSSVersions    map[string]bool      `json:"cvss_versions"`
	Environment     string               `json:"environment"`
	TestingType     string               `json:"testing_type"`
	OSSTMMVector    string               `json:"osstmm_vector"`
	CustomerID      string               `json:"customer_id"`
}

func (d *Driver) AddAssessment(c *fiber.Ctx) error {
	user := c.Locals("user").(*mongo.User)

	// parse request body
	data := &assessmentRequestData{}
	if err := c.BodyParser(data); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// check if user has access to customer
	customer, errStr := d.customerFromParam(data.CustomerID)
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
	targets := []mongo.Target{}
	for _, target := range data.Targets {
		targetID, err := util.ParseUUID(target)
		if err != nil {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": "Invalid target ID",
			})
		}
		targets = append(targets, mongo.Target{
			Model: mongo.Model{
				ID: targetID,
			},
		})
	}

	assessment := &mongo.Assessment{
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
	assessmentID, err := d.mongo.Assessment().Insert(context.Background(), assessment, customer.ID)
	if err != nil {
		c.Status(fiber.StatusBadRequest)

		if mongo.IsDuplicateKeyError(err) {
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
	user := c.Locals("user").(*mongo.User)

	// parse query parameters
	customerParam := c.Query("customer")
	nameParam := c.Query("name")

	var customerID uuid.UUID
	if customerParam != "" {
		// check if user can access the customer
		customer, errStr := d.customerFromParam(customerParam)
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
	if user.Role == mongo.RoleAdmin {
		customers = nil
	}

	// retrieve assessments
	assessments, err := d.mongo.Assessment().Search(context.Background(), customers, customerID, nameParam)
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
	user := c.Locals("user").(*mongo.User)

	// check if user can access the customer
	customer, errStr := d.customerFromParam(c.Params("customer"))
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
	assessments, err := d.mongo.Assessment().GetByCustomerID(context.Background(), customer.ID)
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
	user := c.Locals("user").(*mongo.User)
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

	assessment, err := d.mongo.Assessment().GetByIDPipeline(context.Background(), assessmentID)
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
	user := c.Locals("user").(*mongo.User)

	if len(user.Assessments) == 0 {
		c.Status(fiber.StatusOK)
		return c.JSON([]mongo.Assessment{})
	}

	// map user assessments IDs
	userAssessments := make([]uuid.UUID, len(user.Assessments))
	for i, userAssessment := range user.Assessments {
		userAssessments[i] = userAssessment.ID
	}

	// get assessments from database
	assessments, err := d.mongo.Assessment().GetMultipleByID(context.Background(), userAssessments)
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
	user := c.Locals("user").(*mongo.User)

	// parse assessment param
	assessment, errStr := d.assessmentFromParam(c.Params("assessment"))
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
	targets := []mongo.Target{}
	for _, target := range data.Targets {
		targetID, err := util.ParseUUID(target)
		if err != nil {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": "Invalid target ID",
			})
		}
		targets = append(targets, mongo.Target{
			Model: mongo.Model{
				ID: targetID,
			},
		})
	}

	newAssessment := &mongo.Assessment{
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
	err := d.mongo.Assessment().Update(context.Background(), assessment.ID, newAssessment)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)

		if mongo.IsDuplicateKeyError(err) {
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
	user := c.Locals("user").(*mongo.User)

	// parse assessment param
	assessment, errStr := d.assessmentFromParam(c.Params("assessment"))
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

	newAssessmentStatus := &mongo.Assessment{
		Status: data.Status,
	}

	// update assessment in database
	err := d.mongo.Assessment().UpdateStatus(context.Background(), assessment.ID, newAssessmentStatus)
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
	user := c.Locals("user").(*mongo.User)

	// parse assessment param
	assessment, errStr := d.assessmentFromParam(c.Params("assessment"))
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

	session, err := d.mongo.NewSession()
	if err != nil {
		return err
	}
	defer session.End()

	_, err = session.WithTransaction(func(ctx context.Context) (any, error) {
		// delete assessment from database
		err := d.mongo.Assessment().Delete(ctx, assessment.ID)
		if err != nil {
			return nil, errors.New("Cannot delete assessment")
		}

		return nil, nil
	})
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Assessment deleted",
	})
}

func (d *Driver) CloneAssessment(c *fiber.Ctx) error {
	user := c.Locals("user").(*mongo.User)

	// parse assessment param
	assessment, errStr := d.assessmentFromParam(c.Params("assessment"))
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

	session, err := d.mongo.NewSession()
	if err != nil {
		return err
	}
	defer session.End()

	cloneAssessmentID, err := session.WithTransaction(func(ctx context.Context) (any, error) {
		// clone assessment
		cloneAssessmentID, err := d.mongo.Assessment().Clone(ctx, assessment.ID, data.Name, data.IncludePocs)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
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
	user := c.Locals("user").(*mongo.User)

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

	assessment, err := d.mongo.Assessment().GetByIDPipeline(context.Background(), assessmentID)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Invalid Assessment ID",
		})
	}

	// check if user has access to customer
	customer, err := d.mongo.Customer().GetByID(context.Background(), assessment.Customer.ID)
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
		logoData, _, err := d.mongo.FileReference().ReadByID(context.Background(), customer.LogoID)
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
		template, errStr := d.templateFromParam(data.Template)
		if errStr != "" {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": errStr,
			})
		}

		// retrieve template from database
		templateBytes, _, err = d.mongo.FileReference().ReadByID(context.Background(), template.FileID)
		if err != nil {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": "Invalid template ID",
			})
		}
	}

	// retrieve vulnerabilities
	vulnerabilities, err := d.mongo.Vulnerability().GetByAssessmentIDPocPipeline(context.Background(), assessment.ID, data.IncludeInformationalVulnerabilities)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Failed to retrieve vulnerabilities",
		})
	}

	// retrieve pocs and aggregate vulnerabilities
	for i := range vulnerabilities {
		// TODO: make a function for this or move in the database vulnerability retrieval
		category, err := d.mongo.Category().GetByID(context.Background(), vulnerabilities[i].Category.ID)
		if err != nil {
			c.Status(fiber.StatusInternalServerError)
			return c.JSON(fiber.Map{
				"error": "Cannot get category",
			})
		}

		if vulnerabilities[i].GenericDescription.Enabled {
			vulnerabilities[i].GenericDescription.Text = category.GenericDescription[assessment.Language]
		}

		if vulnerabilities[i].GenericRemediation.Enabled {
			vulnerabilities[i].GenericRemediation.Text = category.GenericRemediation[assessment.Language]
		}

		for j, item := range vulnerabilities[i].Poc.Pocs {
			if item.ImageID != uuid.Nil {
				imageData, _, err := d.mongo.FileReference().ReadByID(context.Background(), item.ImageID)
				if err != nil {
					c.Status(fiber.StatusInternalServerError)
					return c.JSON(fiber.Map{
						"error": "Failed to read image data",
					})
				}
				vulnerabilities[i].Poc.Pocs[j].ImageData = imageData
			}
		}
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

func (d *Driver) assessmentFromParam(assessmentParam string) (*mongo.Assessment, string) {
	if assessmentParam == "" {
		return nil, "Assessment ID is required"
	}

	assessmentID, err := util.ParseUUID(assessmentParam)
	if err != nil {
		return nil, "Invalid assessment ID"
	}

	assessment, err := d.mongo.Assessment().GetByID(context.Background(), assessmentID)
	if err != nil {
		return nil, "Invalid assessment ID"
	}

	return assessment, ""
}

func (d *Driver) validateAssessmentStatus(data *assessmentRequestData) string {
	switch data.Status {
	case
		mongo.ASSESSMENT_STATUS_ON_HOLD,
		mongo.ASSESSMENT_STATUS_IN_PROGRESS,
		mongo.ASSESSMENT_STATUS_COMPLETED:
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
