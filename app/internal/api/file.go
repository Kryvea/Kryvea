package api

import (
	"bytes"
	"context"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (d *Driver) GetImage(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	imageRef, errStr := d.fileFromParam(c.UserContext(), c.Params("file"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	// retrieve the pocs referencing the image
	pocs, err := d.db.Poc().GetByImageID(c.UserContext(), imageRef.ID)
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot retrieve POCs")
	}

	canAccess := false
	for _, poc := range pocs {
		vulnerability, err := d.db.Vulnerability().GetByID(c.UserContext(), poc.VulnerabilityID)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, "Cannot retrieve vulnerability")
		}

		if user.CanAccessCustomer(vulnerability.Customer.ID) {
			canAccess = true
			break
		}
	}
	if !canAccess {
		return jsonError(c, fiber.StatusForbidden, "Forbidden")
	}

	imageData, fileReference, err := d.db.FileReference().ReadByID(c.UserContext(), imageRef.ID)
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot retrieve image")
	}

	c.Status(fiber.StatusOK)
	c.Set("Content-Type", fileReference.MimeType)
	return c.SendStream(bytes.NewReader(imageData))
}

func (d *Driver) GetTemplateFile(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	fileRef, errStr := d.fileFromParam(c.UserContext(), c.Params("file"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	template, err := d.db.Template().GetByFileID(c.UserContext(), fileRef.ID)
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot retrieve template")
	}

	if template.Customer.ID != uuid.Nil && !user.CanAccessCustomer(template.Customer.ID) {
		return jsonError(c, fiber.StatusForbidden, "Forbidden")
	}

	fileData, fileReference, err := d.db.FileReference().ReadByID(c.UserContext(), fileRef.ID)
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot retrieve file")
	}

	c.Status(fiber.StatusOK)
	c.Set("Content-Type", fileReference.MimeType)
	c.Set("Content-Disposition", util.ContentDispositionAttachment(template.Filename))
	return c.SendStream(bytes.NewReader(fileData))
}

func (d *Driver) GetCustomerImage(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	imageRef, errStr := d.fileFromParam(c.UserContext(), c.Params("file"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	canAccessCustomer := false
	for _, usedBy := range imageRef.UsedBy {
		if user.CanAccessCustomer(usedBy) {
			canAccessCustomer = true
			break
		}
	}

	if !canAccessCustomer {
		return jsonError(c, fiber.StatusForbidden, "Forbidden")
	}

	imageData, fileReference, err := d.db.FileReference().ReadByID(c.UserContext(), imageRef.ID)
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot retrieve image")
	}

	c.Status(fiber.StatusOK)
	c.Set("Content-Type", fileReference.MimeType)
	return c.SendStream(bytes.NewReader(imageData))
}

func (d *Driver) fileFromParam(ctx context.Context, param string) (*model.FileReference, string) {
	return fromParam(ctx, param, "File", d.db.FileReference().GetByID)
}
