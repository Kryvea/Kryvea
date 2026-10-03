package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type userRequestData struct {
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Role      string   `json:"role"`
	Customers []string `json:"customers"`
}

func (d *Driver) AddUser(c *fiber.Ctx) error {
	data := &userRequestData{}
	if err := c.BodyParser(data); err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	data.Username = strings.TrimSpace(data.Username)

	errStr := d.validateUserData(data)
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	customers, errStr := d.customersFromIDs(c.UserContext(), data.Customers)
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	user := &model.User{
		Username:  data.Username,
		Role:      data.Role,
		Customers: customers,
	}

	userID, err := d.db.User().Insert(c.UserContext(), user, data.Password)
	if err != nil {
		if errors.Is(err, store.ErrDuplicateKey) {
			return jsonError(c, fiber.StatusBadRequest, fmt.Sprintf("User \"%s\" already exists", user.Username))
		}

		return jsonError(c, fiber.StatusInternalServerError, "Cannot add user")
	}

	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"message": "User added successfully",
		"user_id": userID,
	})
}

type loginRequestData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (d *Driver) Login(c *fiber.Ctx) error {
	data := &loginRequestData{}
	if err := c.BodyParser(data); err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	if data.Username == "" {
		return jsonError(c, fiber.StatusBadRequest, "Username is required")
	}

	if data.Password == "" {
		return jsonError(c, fiber.StatusBadRequest, "Password is required")
	}

	user, err := d.db.User().Login(c.UserContext(), data.Username, data.Password)
	if err != nil {
		if err == store.ErrDisabledUser {
			return jsonError(c, fiber.StatusUnauthorized, "User is disabled")
		}

		return jsonError(c, fiber.StatusUnauthorized, "Invalid credentials")
	}

	if user.PasswordExpiry.Before(time.Now()) {
		util.SetKryveaCookie(c, user.Token.String(), user.TokenExpiry)
		util.SetKryveaShadowCookie(c, util.CookiePasswordExpired, user.TokenExpiry)

		return jsonError(c, fiber.StatusForbidden, "Password expired")
	}

	c.Locals("user", user)
	util.SetSessionCookies(c, user.Role, user.Token, user.TokenExpiry)

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "User logged in successfully",
	})
}

func (d *Driver) GetUsers(c *fiber.Ctx) error {
	users, err := d.db.User().GetAll(c.UserContext())
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot get users")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(users)
}

func (d *Driver) GetUsernames(c *fiber.Ctx) error {
	usernames, err := d.db.User().GetAllUsernames(c.UserContext())
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot get usernames")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(usernames)
}

func (d *Driver) GetMe(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	userData, err := d.db.User().Get(c.UserContext(), user.ID)
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot get user")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(userData)
}

func (d *Driver) GetUser(c *fiber.Ctx) error {
	user, errStr := d.userFromParam(c.UserContext(), c.Params("user"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	c.Status(fiber.StatusOK)
	return c.JSON(user)
}

type updateUserRequestData struct {
	DisabledAt time.Time `json:"disabled_at"`
	Username   string    `json:"username"`
	Role       string    `json:"role"`
	Customers  []string  `json:"customers"`
}

func (d *Driver) UpdateUser(c *fiber.Ctx) error {
	currentUser := c.Locals("user").(*model.User)

	user, errStr := d.userFromParam(c.UserContext(), c.Params("user"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	if currentUser.ID == user.ID {
		return jsonError(c, fiber.StatusBadRequest, "Cannot update self, use dedicated endpoint")
	}

	data := &updateUserRequestData{}
	if err := c.BodyParser(data); err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	errStr = d.validateUserUpdateData(data)
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	customers, errStr := d.customersFromIDs(c.UserContext(), data.Customers)
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	newUser := &model.User{
		DisabledAt: data.DisabledAt,
		Username:   data.Username,
		Role:       data.Role,
		Customers:  customers,
	}

	_, err := d.db.RunInTxWithLock(c.UserContext(), model.LockAdmin, func(ctx context.Context) (any, error) {
		err := d.db.User().Update(ctx, user.ID, newUser)
		if err != nil {
			if errors.Is(err, store.ErrDuplicateKey) {
				return nil, fmt.Errorf("User \"%s\" already exists", newUser.Username)
			}

			return nil, errors.New("Cannot update user")
		}

		return nil, nil
	})
	if err != nil {
		if errors.Is(err, store.ErrLocked) {
			err = errors.New("Someone else is currently editing users, retry")
		}
		return jsonError(c, fiber.StatusBadRequest, err.Error())
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "User updated",
	})
}

type updateMeData struct {
	Username        string `json:"username"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (d *Driver) UpdateMe(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	data := &updateMeData{}
	if err := c.BodyParser(data); err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	errStr := d.validateUpdateMeData(c.UserContext(), data, user)
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	newUser := &model.User{
		Username: data.Username,
	}

	_, err := d.db.RunInTxWithLock(c.UserContext(), model.LockUsername, func(ctx context.Context) (any, error) {
		err := d.db.User().UpdateMe(ctx, user.ID, newUser, data.NewPassword)
		if err != nil {
			if errors.Is(err, store.ErrDuplicateKey) {
				return nil, fmt.Errorf("User \"%s\" already exists", newUser.Username)
			}

			return nil, errors.New("Cannot update user")
		}

		return nil, nil
	})
	if err != nil {
		if errors.Is(err, store.ErrLocked) {
			err = errors.New("Someone else is currently editing users, retry")
		}
		return jsonError(c, fiber.StatusBadRequest, err.Error())
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "User updated",
	})
}

func (d *Driver) UpdateOwnedAssessment(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	type reqData struct {
		Assessment string `json:"assessment"`
		IsOwned    bool   `json:"is_owned"`
	}
	data := &reqData{}
	if err := c.BodyParser(data); err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	assessment, errStr := d.assessmentFromParam(c.UserContext(), data.Assessment)
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	if !user.CanAccessCustomer(assessment.Customer.ID) {
		return jsonError(c, fiber.StatusForbidden, "Forbidden")
	}

	err := d.db.User().UpdateOwnedAssessment(c.UserContext(), user.ID, assessment.ID, data.IsOwned)
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot edit owned assessment")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Owned assessment updated",
	})
}

func (d *Driver) DeleteUser(c *fiber.Ctx) error {
	userID, err := util.ParseUUID(c.Params("user"))
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	_, err = d.db.RunInTxWithLock(c.UserContext(), model.LockAdmin, func(ctx context.Context) (any, error) {
		return nil, d.db.User().Delete(ctx, userID)
	})
	if err != nil {
		if errors.Is(err, store.ErrAdminUserRequired) {
			return jsonError(c, fiber.StatusBadRequest, err.Error())
		}
		if errors.Is(err, store.ErrLocked) {
			return jsonError(c, fiber.StatusBadRequest, "Someone else is currently editing users, retry")
		}
		return jsonError(c, fiber.StatusBadRequest, "Cannot delete user")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "User deleted",
	})
}

func (d *Driver) Logout(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	err := d.db.User().Logout(c.UserContext(), user.ID)
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot logout user")
	}

	util.ClearCookies(c)

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "User logged out",
	})
}

func (d *Driver) ResetUserPassword(c *fiber.Ctx) error {
	user, errStr := d.userFromParam(c.UserContext(), c.Params("user"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	newPassword, err := util.GenerateRandomPassword(20)
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot generate new password")
	}

	err = d.db.User().ResetUserPassword(c.UserContext(), user.ID, newPassword)
	if err != nil {
		return jsonError(c, fiber.StatusInternalServerError, "Cannot reset password")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message":  "Password reset successfully",
		"password": newPassword,
	})
}

func (d *Driver) ResetPassword(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	type reqData struct {
		Password string `json:"password"`
	}
	data := &reqData{}
	if err := c.BodyParser(data); err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	if data.Password == "" {
		return jsonError(c, fiber.StatusBadRequest, "Password is required")
	}

	if !util.IsValidPassword(data.Password) {
		return jsonError(c, fiber.StatusBadRequest, "Password does not meet policy requirements")
	}

	err := d.db.User().ResetPassword(c.UserContext(), user, data.Password)
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot reset password")
	}

	util.SetSessionCookies(c, user.Role, user.Token, user.TokenExpiry)

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "Password reset",
	})
}

func (d *Driver) userFromParam(ctx context.Context, userParam string) (*model.User, string) {
	return fromParam(ctx, userParam, "User", d.db.User().Get)
}

// customersFromIDs validates that every customer ID exists using a single
// bulk lookup and returns them as model.Customer references.
func (d *Driver) customersFromIDs(ctx context.Context, ids []string) ([]model.Customer, string) {
	customers := make([]model.Customer, len(ids))
	if len(ids) == 0 {
		return customers, ""
	}

	customerIDs := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		parsed, err := util.ParseUUID(id)
		if err != nil {
			return nil, "Invalid customer ID"
		}
		customerIDs[i] = parsed
	}

	existing, err := d.db.Customer().ExistingIDs(ctx, customerIDs)
	if err != nil {
		return nil, "Invalid customer ID"
	}

	found := make(map[uuid.UUID]struct{}, len(existing))
	for _, id := range existing {
		found[id] = struct{}{}
	}

	for i, id := range customerIDs {
		if _, ok := found[id]; !ok {
			return nil, "Invalid customer ID"
		}
		customers[i] = model.Customer{
			Model: model.Model{
				ID: id,
			},
		}
	}

	return customers, ""
}

func (d *Driver) validateUserData(data *userRequestData) string {
	if data.Username == "" {
		return "Username is required"
	}

	if !model.IsValidRole(data.Role) {
		return "Invalid role"
	}

	if !util.IsValidPassword(data.Password) {
		return "Password does not meet policy requirements"
	}

	return ""
}

func (d *Driver) validateUserUpdateData(data *updateUserRequestData) string {
	if !model.IsValidRole(data.Role) {
		return "Invalid role"
	}

	return ""
}

func (d *Driver) validateUpdateMeData(ctx context.Context, data *updateMeData, user *model.User) string {
	if data.Username == "" && data.NewPassword == "" {
		return "No data to update"
	}

	if data.NewPassword != "" {
		if data.CurrentPassword == "" {
			return "Current password is required"
		}

		if data.NewPassword == data.CurrentPassword {
			return "New password cannot be the same as current password"
		}

		err := d.db.User().ValidatePassword(ctx, user.ID, data.CurrentPassword)
		if err != nil || !util.IsValidPassword(data.NewPassword) {
			return "Invalid passwords"
		}
	}

	return ""
}
