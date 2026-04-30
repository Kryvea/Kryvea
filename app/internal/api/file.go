package api

import (
	"bytes"
	"context"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (d *Driver) GetImage(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// parse image param
	imageRef, errStr := d.fileFromParam(c.UserContext(), c.Params("file"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// retrieve vulnerability from database
	pocs, err := d.db.Poc().GetByImageID(c.UserContext(), imageRef.ID)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot retrieve POCs",
		})
	}

	canAccess := false
	for _, poc := range pocs {
		vulnerability, err := d.db.Vulnerability().GetByID(c.UserContext(), poc.VulnerabilityID)
		if err != nil {
			c.Status(fiber.StatusInternalServerError)
			return c.JSON(fiber.Map{
				"error": "Cannot retrieve vulnerability",
			})
		}

		assessment, err := d.db.Assessment().GetByID(c.UserContext(), vulnerability.Assessment.ID)
		if err != nil {
			c.Status(fiber.StatusInternalServerError)
			return c.JSON(fiber.Map{
				"error": "Cannot retrieve assessment",
			})
		}

		if user.CanAccessCustomer(assessment.Customer.ID) {
			canAccess = true
			break
		}
	}
	if !canAccess {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	imageData, fileReference, err := d.db.FileReference().ReadByID(c.UserContext(), imageRef.ID)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot retrieve image",
		})
	}

	c.Status(fiber.StatusOK)
	c.Set("Content-Type", fileReference.MimeType)
	return c.SendStream(bytes.NewReader(imageData))
}

func (d *Driver) GetTemplateFile(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// parse image param
	fileRef, errStr := d.fileFromParam(c.UserContext(), c.Params("file"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	template, err := d.db.Template().GetByFileID(c.UserContext(), fileRef.ID)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot retrieve template",
		})
	}

	if template.Customer.ID != uuid.Nil && !user.CanAccessCustomer(template.Customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	fileData, fileReference, err := d.db.FileReference().ReadByID(c.UserContext(), fileRef.ID)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot retrieve file",
		})
	}

	c.Status(fiber.StatusOK)
	c.Set("Content-Type", fileReference.MimeType)
	c.Set("Content-Disposition", "attachment; filename="+template.Filename)
	return c.SendStream(bytes.NewReader(fileData))
}

func (d *Driver) GetCustomerImage(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// parse image param
	imageRef, errStr := d.fileFromParam(c.UserContext(), c.Params("file"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// check if user has access to customer
	canAccessCustomer := false
	for _, usedBy := range imageRef.UsedBy {
		if user.CanAccessCustomer(usedBy) {
			canAccessCustomer = true
			break
		}
	}

	if !canAccessCustomer {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	// read the image from the database
	imageData, fileReference, err := d.db.FileReference().ReadByID(c.UserContext(), imageRef.ID)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot retrieve image",
		})
	}

	c.Status(fiber.StatusOK)
	c.Set("Content-Type", fileReference.MimeType)
	return c.SendStream(bytes.NewReader(imageData))
}

func (d *Driver) fileFromParam(ctx context.Context, param string) (*model.FileReference, string) {
	if param == "" {
		return nil, "File ID is required"
	}

	fileID, err := uuid.Parse(param)
	if err != nil {
		return nil, "Invalid file ID"
	}

	file, err := d.db.FileReference().GetByID(ctx, fileID)
	if err != nil {
		return nil, "Invalid file ID"
	}

	return file, ""
}
