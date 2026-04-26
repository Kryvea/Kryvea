package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kryvea/Kryvea/internal/mongo"
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

	session, err := d.mongo.NewSession()
	if err != nil {
		return err
	}
	defer session.End()

	customerID, err := session.WithTransaction(func(ctx context.Context) (any, error) {
		var logoId uuid.UUID
		var mime string
		logoData, _, err := d.formDataReadImage(c, context.Background(), "file")
		if err == mongo.ErrFileSizeTooLarge {
			return uuid.Nil, errors.New("Image file size is too large")
		}
		if err == mongo.ErrImageTypeNotAllowed {
			return uuid.Nil, errors.New("Image type is not allowed")
		}
		if len(logoData) > 0 && err == nil {
			logoId, mime, err = d.mongo.FileReference().Insert(ctx, logoData)
			if err != nil {
				d.logger.Error().Err(err).Msg("Cannot upload image")
				return uuid.Nil, errors.New("Cannot upload image")
			}
		}

		customer := &mongo.Customer{
			Name:         data.Name,
			Language:     data.Language,
			LogoID:       logoId,
			LogoMimeType: mime,
		}

		// insert customer into database
		customerID, err := d.mongo.Customer().Insert(ctx, customer)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return uuid.Nil, fmt.Errorf("Customer \"%s\" already exists", customer.Name)
			}

			return uuid.Nil, errors.New("Cannot create customer")
		}

		for _, assignedUserStr := range data.AssignedUsers {
			assignedUser, errStr := d.userFromParam(assignedUserStr)
			if errStr != "" {
				return nil, errors.New(errStr)
			}

			err := d.mongo.User().AssignCustomer(ctx, assignedUser.ID, customerID)
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
	user := c.Locals("user").(*mongo.User)

	// get customer from param
	customer, errStr := d.customerFromParam(c.Params("customer"))
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
	user := c.Locals("user").(*mongo.User)

	// retrieve user's customers
	userCustomers := []uuid.UUID{}
	for _, uc := range user.Customers {
		userCustomers = append(userCustomers, uc.ID)
	}
	if user.Role == mongo.RoleAdmin {
		userCustomers = nil
	}

	// get customers from database
	customers, err := d.mongo.Customer().GetAll(context.Background(), userCustomers)
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
	customer, errStr := d.customerFromParam(c.Params("customer"))
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

	newCustomer := &mongo.Customer{
		Name:     data.Name,
		Language: data.Language,
	}

	// update customer in database
	err := d.mongo.Customer().Update(context.Background(), customer.ID, newCustomer)
	if err != nil {
		c.Status(fiber.StatusBadRequest)

		if mongo.IsDuplicateKeyError(err) {
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
	customer, errStr := d.customerFromParam(c.Params("customer"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	session, err := d.mongo.NewSession()
	if err != nil {
		return err
	}
	defer session.End()

	_, err = session.WithTransaction(func(ctx context.Context) (any, error) {
		logoData, _, err := d.formDataReadImage(c, context.Background(), "file")
		if err != nil {
			if err == mongo.ErrFileSizeTooLarge {
				return uuid.Nil, errors.New("Image file size is too large")
			}
			if err == mongo.ErrImageTypeNotAllowed {
				return uuid.Nil, errors.New("Image type is not allowed")
			}

			return uuid.Nil, errors.New("Error reading form file")
		}

		var logoId uuid.UUID
		var mime string
		if len(logoData) > 0 {
			logoId, mime, err = d.mongo.FileReference().Insert(ctx, logoData)
			if err != nil {
				return nil, errors.New("Cannot upload image")
			}
		}

		// update customer in database
		err = d.mongo.Customer().UpdateLogo(ctx, customer.ID, logoId, mime)
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

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Customer logo updated",
	})
}

func (d *Driver) DeleteCustomer(c *fiber.Ctx) error {
	// parse customer param
	customer, errStr := d.customerFromParam(c.Params("customer"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	session, err := d.mongo.NewSession()
	if err != nil {
		return err
	}
	defer session.End()

	_, err = session.WithTransaction(func(ctx context.Context) (any, error) {
		err := d.mongo.Customer().Delete(ctx, customer.ID)
		if err != nil {
			d.logger.Error().Err(err).Msg("Cannot delete customer")
			return nil, errors.New("Cannot delete customer")
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
		"message": "Customer deleted",
	})
}

func (d *Driver) customerFromParam(customerParam string) (*mongo.Customer, string) {
	if customerParam == "" {
		return nil, "Customer ID is required"
	}

	customerID, err := util.ParseUUID(customerParam)
	if err != nil {
		return nil, "Invalid customer ID"
	}

	customer, err := d.mongo.Customer().GetByIDPipeline(context.Background(), customerID)
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
