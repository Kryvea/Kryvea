package api

import (
	"context"
	"errors"

	"github.com/Kryvea/Kryvea/internal/mongo"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type targetRequestData struct {
	IPv4         string `json:"ipv4"`
	IPv6         string `json:"ipv6"`
	Port         int    `json:"port"`
	Protocol     string `json:"protocol"`
	FQDN         string `json:"fqdn"`
	Tag          string `json:"tag"`
	CustomerID   string `json:"customer_id"`
	AssessmentID string `json:"assessment_id"`
}

func (d *Driver) AddTarget(c *fiber.Ctx) error {
	user := c.Locals("user").(*mongo.User)

	// parse request body
	data := &targetRequestData{}
	if err := c.BodyParser(data); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// validate data
	errStr := d.validateTargetData(data)
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
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

	assessment := &mongo.Assessment{
		Model: mongo.Model{
			ID: uuid.Nil,
		},
	}
	// if assessment is not empty retrieve it from database
	if data.AssessmentID != "" {
		assessment, errStr = d.assessmentFromParam(data.AssessmentID)
		if errStr != "" {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": errStr,
			})
		}
	}

	target := &mongo.Target{
		IPv4:     data.IPv4,
		IPv6:     data.IPv6,
		Port:     data.Port,
		Protocol: data.Protocol,
		FQDN:     data.FQDN,
		Tag:      data.Tag,
	}

	session, err := d.mongo.NewSession()
	if err != nil {
		return err
	}
	defer session.End()

	targetID, err := session.WithTransaction(func(ctx context.Context) (any, error) {
		// insert target into database
		targetID, err := d.mongo.Target().Insert(ctx, target, customer.ID)
		if err != nil {
			c.Status(fiber.StatusBadRequest)

			if mongo.IsDuplicateKeyError(err) {
				return uuid.Nil, errors.New("Target with provided data already exists")
			}

			return uuid.Nil, errors.New("Cannot create target")
		}

		// add target to assessment if provided
		if assessment.ID != uuid.Nil {
			err = d.mongo.Assessment().UpdateTargets(ctx, assessment.ID, target.ID)
			if err != nil {
				return uuid.Nil, errors.New("Cannot add target to assessment")
			}
		}

		return targetID, nil
	})
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"message":   "Target created",
		"target_id": targetID.(uuid.UUID),
	})
}

func (d *Driver) UpdateTarget(c *fiber.Ctx) error {
	user := c.Locals("user").(*mongo.User)

	// parse target param
	target, errStr := d.targetFromParam(c.Params("target"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// check if user has access to customer
	if !user.CanAccessCustomer(target.Customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	// parse request body
	data := &targetRequestData{}
	if err := c.BodyParser(data); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// validate data
	errStr = d.validateTargetData(data)
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	newTarget := &mongo.Target{
		IPv4:     data.IPv4,
		IPv6:     data.IPv6,
		Port:     data.Port,
		Protocol: data.Protocol,
		FQDN:     data.FQDN,
		Tag:      data.Tag,
	}

	// update target in database
	err := d.mongo.Target().Update(context.Background(), target.ID, newTarget)
	if err != nil {
		c.Status(fiber.StatusBadRequest)

		if mongo.IsDuplicateKeyError(err) {
			return c.JSON(fiber.Map{
				"error": "Target with provided data already exists",
			})
		}

		return c.JSON(fiber.Map{
			"error": "Cannot update target",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Target updated",
	})
}

func (d *Driver) DeleteTarget(c *fiber.Ctx) error {
	user := c.Locals("user").(*mongo.User)

	// parse target param
	target, errStr := d.targetFromParam(c.Params("target"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// check if user has access to customer
	if !user.CanAccessCustomer(target.Customer.ID) {
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
		// delete target from database
		err := d.mongo.Target().Delete(ctx, target.ID)
		if err != nil {
			return nil, errors.New("Cannot delete target")
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
		"message": "Target deleted",
	})
}

func (d *Driver) GetTargetsByCustomer(c *fiber.Ctx) error {
	user := c.Locals("user").(*mongo.User)

	// check if user has access to customer
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

	targets, err := d.mongo.Target().Search(context.Background(), customer.ID, c.Query("search"))
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot get targets",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(targets)
}

func (d *Driver) GetTarget(c *fiber.Ctx) error {
	user := c.Locals("user").(*mongo.User)

	// parse target param
	targetParam := c.Params("target")
	if targetParam == "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Target ID is required",
		})
	}

	targetID, err := util.ParseUUID(targetParam)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Invalid target ID",
		})
	}

	// get target by customer and ID from database
	target, err := d.mongo.Target().GetByIDPipeline(context.Background(), targetID)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot get target",
		})
	}

	if !user.CanAccessCustomer(target.Customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(target)
}

func (d *Driver) targetFromParam(targetParam string) (*mongo.Target, string) {
	if targetParam == "" {
		return nil, "Target ID is required"
	}

	targetID, err := util.ParseUUID(targetParam)
	if err != nil {
		return nil, "Invalid target ID"
	}

	target, err := d.mongo.Target().GetByIDPipeline(context.Background(), targetID)
	if err != nil {
		return nil, "Invalid target ID"
	}

	return target, ""
}

func (d *Driver) validateTargetData(data *targetRequestData) string {
	if data.FQDN == "" && data.IPv4 == "" && data.IPv6 == "" {
		return "At least one of FQDN/Target name, IPv4 or IPv6 must be provided"
	}

	if data.IPv4 != "" && !util.IsValidIPv4(data.IPv4) {
		return "Invalid IPv4 address"
	}

	if data.IPv6 != "" && !util.IsValidIPv6(data.IPv6) {
		return "Invalid IPv6 address"
	}

	return ""
}
