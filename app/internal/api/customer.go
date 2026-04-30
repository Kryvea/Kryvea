package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type customerRequestData struct {
	Name          string   `json:"name"`
	Language      string   `json:"language"`
	AssignedUsers []string `json:"assigned_users"`
}

func (d *Driver) AddCustomer(c *fiber.Ctx) error {
	// parse request body
	data := &customerRequestData{}
	dataStr := c.FormValue("data")
	err := sonic.Unmarshal([]byte(dataStr), &data)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// validate data
	errStr := d.validateCustomerData(data)
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	customerID, err := d.db.RunInTx(c.UserContext(), func(ctx context.Context) (any, error) {
		var logoId uuid.UUID
		var mime string
		logoData, _, err := d.formDataReadImage(c, ctx, "file")
		if err == store.ErrFileSizeTooLarge {
			return uuid.Nil, errors.New("Image file size is too large")
		}
		if err == store.ErrImageTypeNotAllowed {
			return uuid.Nil, errors.New("Image type is not allowed")
		}
		if len(logoData) > 0 && err == nil {
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

		// insert customer into database
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
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"message":     "Customer created",
		"customer_id": customerID.(uuid.UUID),
	})
}

func (d *Driver) GetCustomer(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// get customer from param
	customer, errStr := d.customerFromParam(c.UserContext(), c.Params("customer"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// check if user has access to the customer
	if !user.CanAccessCustomer(customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(customer)
}

func (d *Driver) GetCustomers(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// retrieve user's customers
	userCustomers := []uuid.UUID{}
	for _, uc := range user.Customers {
		userCustomers = append(userCustomers, uc.ID)
	}
	if user.Role == model.RoleAdmin {
		userCustomers = nil
	}

	// get customers from database
	customers, err := d.db.Customer().GetAll(c.UserContext(), userCustomers)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot get customers",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(customers)
}

func (d *Driver) UpdateCustomer(c *fiber.Ctx) error {
	// parse customer param
	customer, errStr := d.customerFromParam(c.UserContext(), c.Params("customer"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// parse request body
	data := &customerRequestData{}
	if err := c.BodyParser(data); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// validate data
	errStr = d.validateCustomerData(data)
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	newCustomer := &model.Customer{
		Name:     data.Name,
		Language: data.Language,
	}

	// update customer in database
	err := d.db.Customer().Update(c.UserContext(), customer.ID, newCustomer)
	if err != nil {
		c.Status(fiber.StatusBadRequest)

		if errors.Is(err, store.ErrDuplicateKey) {
			return c.JSON(fiber.Map{
				"error": fmt.Sprintf("Customer \"%s\" already exists", newCustomer.Name),
			})
		}

		return c.JSON(fiber.Map{
			"error": "Cannot update customer",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Customer updated",
	})
}

func (d *Driver) UpdateCustomerLogo(c *fiber.Ctx) error {
	// parse customer param
	customer, errStr := d.customerFromParam(c.UserContext(), c.Params("customer"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	_, err := d.db.RunInTx(c.UserContext(), func(ctx context.Context) (any, error) {
		logoData, _, err := d.formDataReadImage(c, ctx, "file")
		if err != nil {
			if err == store.ErrFileSizeTooLarge {
				return uuid.Nil, errors.New("Image file size is too large")
			}
			if err == store.ErrImageTypeNotAllowed {
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

		// update customer in database
		err = d.db.Customer().UpdateLogo(ctx, customer.ID, logoId, mime)
		if err != nil {
			return nil, errors.New("Cannot update customer")
		}

		return nil, nil
	})
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	d.gcFilesAsync()
	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Customer logo updated",
	})
}

func (d *Driver) DeleteCustomer(c *fiber.Ctx) error {
	// parse customer param
	customer, errStr := d.customerFromParam(c.UserContext(), c.Params("customer"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	if err := d.db.Customer().Delete(c.UserContext(), customer.ID); err != nil {
		d.logger.Error().Err(err).Msg("Cannot delete customer")
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot delete customer",
		})
	}

	d.gcFilesAsync()
	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Customer deleted",
	})
}

func (d *Driver) customerFromParam(ctx context.Context, customerParam string) (*model.Customer, string) {
	if customerParam == "" {
		return nil, "Customer ID is required"
	}

	customerID, err := util.ParseUUID(customerParam)
	if err != nil {
		return nil, "Invalid customer ID"
	}

	customer, err := d.db.Customer().GetByIDWithRelations(ctx, customerID)
	if err != nil {
		return nil, "Invalid customer ID"
	}

	return customer, ""
}

func (d *Driver) validateCustomerData(customer *customerRequestData) string {
	if customer.Name == "" {
		return "Name is required"
	}

	return ""
}
