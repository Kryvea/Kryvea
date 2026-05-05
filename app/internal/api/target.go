package api

import (
	"context"
	"errors"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type targetUpsertItem struct {
	ID       string `json:"id"`
	IPv4     string `json:"ipv4"`
	IPv6     string `json:"ipv6"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	FQDN     string `json:"fqdn"`
	Tag      string `json:"tag"`
}

type targetBulkRequest struct {
	CustomerID   string             `json:"customer_id"`
	AssessmentID string             `json:"assessment_id"`
	Upsert       []targetUpsertItem `json:"upsert"`
	Delete       []string           `json:"delete"`
}

func (d *Driver) BulkTargets(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// parse request body
	data := &targetBulkRequest{}
	if err := c.BodyParser(data); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// check if user has access to customer
	customer, errStr := d.customerFromParam(c.UserContext(), data.CustomerID)
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{"error": errStr})
	}
	if !user.CanAccessCustomer(customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{"error": "Forbidden"})
	}

	// if assessment is not empty retrieve it from database
	assessment := &model.Assessment{Model: model.Model{ID: uuid.Nil}}
	if data.AssessmentID != "" {
		assessment, errStr = d.assessmentFromParam(c.UserContext(), data.AssessmentID)
		if errStr != "" {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{"error": errStr})
		}
	}

	// validate data
	for i := range data.Upsert {
		if errStr := d.validateTargetItem(&data.Upsert[i]); errStr != "" {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{"error": errStr})
		}
	}

	deleteIDs := make([]uuid.UUID, 0, len(data.Delete))
	for _, raw := range data.Delete {
		id, err := util.ParseUUID(raw)
		if err != nil {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{"error": "Invalid target ID"})
		}
		deleteIDs = append(deleteIDs, id)
	}

	updateIDs := make([]uuid.UUID, 0, len(data.Upsert))
	for _, item := range data.Upsert {
		if item.ID == "" {
			continue
		}
		id, err := util.ParseUUID(item.ID)
		if err != nil {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{"error": "Invalid target ID"})
		}
		updateIDs = append(updateIDs, id)
	}

	referencedIDs := append(append(make([]uuid.UUID, 0, len(deleteIDs)+len(updateIDs)), deleteIDs...), updateIDs...)
	if len(referencedIDs) > 0 {
		found, err := d.db.Target().ExistingIDsForCustomer(c.UserContext(), referencedIDs, customer.ID)
		if err != nil {
			c.Status(fiber.StatusInternalServerError)
			return c.JSON(fiber.Map{"error": "Cannot validate targets"})
		}
		allowed := make(map[uuid.UUID]struct{}, len(found))
		for _, id := range found {
			allowed[id] = struct{}{}
		}
		for _, id := range referencedIDs {
			if _, ok := allowed[id]; !ok {
				c.Status(fiber.StatusBadRequest)
				return c.JSON(fiber.Map{"error": "Invalid target ID"})
			}
		}
	}

	insertedIDs, err := d.db.RunInTx(c.UserContext(), func(ctx context.Context) (any, error) {
		var inserted []uuid.UUID
		for _, id := range deleteIDs {
			if err := d.db.Target().Delete(ctx, id); err != nil {
				return nil, errors.New("Cannot delete target")
			}
		}
		for _, item := range data.Upsert {
			t := &model.Target{
				IPv4:     item.IPv4,
				IPv6:     item.IPv6,
				Port:     item.Port,
				Protocol: item.Protocol,
				FQDN:     item.FQDN,
				Tag:      item.Tag,
			}
			if item.ID == "" {
				id, err := d.db.Target().Insert(ctx, t, customer.ID)
				if err != nil {
					if errors.Is(err, store.ErrDuplicateKey) {
						return nil, errors.New("Target with provided data already exists")
					}
					return nil, errors.New("Cannot create target")
				}
				inserted = append(inserted, id)
				continue
			}
			id, _ := util.ParseUUID(item.ID)
			if err := d.db.Target().Update(ctx, id, t); err != nil {
				if errors.Is(err, store.ErrDuplicateKey) {
					return nil, errors.New("Target with provided data already exists")
				}
				return nil, errors.New("Cannot update target")
			}
		}
		// add targets to assessment if provided
		if assessment.ID != uuid.Nil && len(inserted) > 0 {
			if err := d.db.Assessment().BulkUpdateTargets(ctx, assessment.ID, inserted); err != nil {
				return nil, errors.New("Cannot attach targets to assessment")
			}
		}
		return inserted, nil
	})
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{"error": err.Error()})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message":      "Targets updated",
		"inserted_ids": insertedIDs,
	})
}

func (d *Driver) GetTargetsByCustomer(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// check if user has access to customer
	customer, errStr := d.customerFromParam(c.UserContext(), c.Params("customer"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{"error": errStr})
	}

	if !user.CanAccessCustomer(customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{"error": "Forbidden"})
	}

	targets, err := d.db.Target().Search(c.UserContext(), customer.ID, c.Query("search"))
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{"error": "Cannot get targets"})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(targets)
}

func (d *Driver) GetTarget(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// parse target param
	targetParam := c.Params("target")
	if targetParam == "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{"error": "Target ID is required"})
	}

	targetID, err := util.ParseUUID(targetParam)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{"error": "Invalid target ID"})
	}

	// get target by customer and ID from database
	target, err := d.db.Target().GetByIDWithRelations(c.UserContext(), targetID)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{"error": "Cannot get target"})
	}

	if !user.CanAccessCustomer(target.Customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{"error": "Forbidden"})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(target)
}

func (d *Driver) targetFromParam(ctx context.Context, targetParam string) (*model.Target, string) {
	if targetParam == "" {
		return nil, "Target ID is required"
	}

	targetID, err := util.ParseUUID(targetParam)
	if err != nil {
		return nil, "Invalid target ID"
	}

	target, err := d.db.Target().GetByIDWithRelations(ctx, targetID)
	if err != nil {
		return nil, "Invalid target ID"
	}

	return target, ""
}

func (d *Driver) validateTargetItem(data *targetUpsertItem) string {
	if data.FQDN == "" && data.IPv4 == "" && data.IPv6 == "" {
		return "At least one of FQDN/Target name, IPv4 or IPv6 must be provided"
	}
	if data.IPv4 != "" && !util.IsValidIPv4(data.IPv4) {
		return "Invalid IPv4 address"
	}
	if data.IPv6 != "" && !util.IsValidIPv6(data.IPv6) {
		return "Invalid IPv6 address"
	}
	if data.Port < 0 || data.Port > 65535 {
		return "Port must be between 0 and 65535"
	}
	return ""
}
