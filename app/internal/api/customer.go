package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

type customerRequestData struct {
	Name          string   `json:"name"`
	Language      string   `json:"language"`
	AssignedUsers []string `json:"assigned_users"`
}

func (d *Driver) AddCustomer(c *fiber.Ctx) error {
	data := &customerRequestData{}
	dataStr := c.FormValue("data")
	err := sonic.Unmarshal([]byte(dataStr), data)
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	errStr := d.validateCustomerData(data)
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	customerID, err := d.db.RunInTx(c.UserContext(), func(ctx context.Context) (any, error) {
		var logoId uuid.UUID
		var mime string
		logoData, _, err := d.formDataReadImage(c, ctx, "file")
		if err != nil && !errors.Is(err, fasthttp.ErrMissingFile) {
			// the logo is optional: a missing file is fine, anything else is reported
			if errors.Is(err, store.ErrFileSizeTooLarge) {
				return uuid.Nil, errors.New("Image file size is too large")
			}
			if errors.Is(err, store.ErrImageTypeNotAllowed) {
				return uuid.Nil, errors.New("Image type is not allowed")
			}

			return uuid.Nil, errors.New("Error reading form file")
		}
		if err == nil && len(logoData) > 0 {
			logoId, mime, err = d.db.FileReference().Insert(ctx, logoData)
			if err != nil {
				d.logger.Error().Err(err).Msg("Cannot upload image")
				return uuid.Nil, errors.New("Cannot upload image")
			}
		}

		customer := &model.Customer{
			Name:         data.Name,
			Language:     data.Language,
			LogoID:       logoId,
			LogoMimeType: mime,
		}

		customerID, err := d.db.Customer().Insert(ctx, customer)
		if err != nil {
			if errors.Is(err, store.ErrDuplicateKey) {
				return uuid.Nil, fmt.Errorf("Customer \"%s\" already exists", customer.Name)
			}

			return uuid.Nil, errors.New("Cannot create customer")
		}

		for _, assignedUserStr := range data.AssignedUsers {
			assignedUser, errStr := d.userFromParam(ctx, assignedUserStr)
			if errStr != "" {
				return nil, errors.New(errStr)
			}

			err := d.db.User().AssignCustomer(ctx, assignedUser.ID, customerID)
			if err != nil {
				return nil, err
			}
		}

		return customerID, nil
	})
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, err.Error())
	}

	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"message":     "Customer created",
		"customer_id": customerID.(uuid.UUID),
	})
}

func (d *Driver) GetCustomer(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	customer, errStr := d.customerFromParam(c.UserContext(), c.Params("customer"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	if !user.CanAccessCustomer(customer.ID) {
		return jsonError(c, fiber.StatusForbidden, "Forbidden")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(customer)
}

func (d *Driver) GetCustomers(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	userCustomers := userCustomerIDs(user)

	customers, err := d.db.Customer().GetAll(c.UserContext(), userCustomers)
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot get customers")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(customers)
}

func (d *Driver) UpdateCustomer(c *fiber.Ctx) error {
	customer, errStr := d.customerFromParam(c.UserContext(), c.Params("customer"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	data := &customerRequestData{}
	if err := c.BodyParser(data); err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	errStr = d.validateCustomerData(data)
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	newCustomer := &model.Customer{
		Name:     data.Name,
		Language: data.Language,
	}

	err := d.db.Customer().Update(c.UserContext(), customer.ID, newCustomer)
	if err != nil {
		if errors.Is(err, store.ErrDuplicateKey) {
			return jsonError(c, fiber.StatusBadRequest, fmt.Sprintf("Customer \"%s\" already exists", newCustomer.Name))
		}

		return jsonError(c, fiber.StatusBadRequest, "Cannot update customer")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Customer updated",
	})
}

func (d *Driver) UpdateCustomerLogo(c *fiber.Ctx) error {
	customer, errStr := d.customerFromParam(c.UserContext(), c.Params("customer"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	_, err := d.db.RunInTx(c.UserContext(), func(ctx context.Context) (any, error) {
		logoData, _, err := d.formDataReadImage(c, ctx, "file")
		if err != nil {
			if errors.Is(err, store.ErrFileSizeTooLarge) {
				return uuid.Nil, errors.New("Image file size is too large")
			}
			if errors.Is(err, store.ErrImageTypeNotAllowed) {
				return uuid.Nil, errors.New("Image type is not allowed")
			}

			return uuid.Nil, errors.New("Error reading form file")
		}

		var logoId uuid.UUID
		var mime string
		if len(logoData) > 0 {
			logoId, mime, err = d.db.FileReference().Insert(ctx, logoData)
			if err != nil {
				return nil, errors.New("Cannot upload image")
			}
		}

		err = d.db.Customer().UpdateLogo(ctx, customer.ID, logoId, mime)
		if err != nil {
			return nil, errors.New("Cannot update customer")
		}

		return nil, nil
	})
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, err.Error())
	}

	d.gcFilesAsync()
	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Customer logo updated",
	})
}

func (d *Driver) DeleteCustomer(c *fiber.Ctx) error {
	customer, errStr := d.customerFromParam(c.UserContext(), c.Params("customer"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	if err := d.db.Customer().Delete(c.UserContext(), customer.ID); err != nil {
		d.logger.Error().Err(err).Msg("Cannot delete customer")
		return jsonError(c, fiber.StatusBadRequest, "Cannot delete customer")
	}

	d.gcFilesAsync()
	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Customer deleted",
	})
}

func (d *Driver) customerFromParam(ctx context.Context, customerParam string) (*model.Customer, string) {
	return fromParam(ctx, customerParam, "Customer", d.db.Customer().GetByIDWithRelations)
}

func (d *Driver) validateCustomerData(customer *customerRequestData) string {
	if customer.Name == "" {
		return "Name is required"
	}

	return ""
}
