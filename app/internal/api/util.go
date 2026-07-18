package api

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"strings"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// trimCutset is the set of characters trimmed from user-provided text fields.
const trimCutset = "\r\n "

// fromParam parses a UUID request parameter and fetches the corresponding
// entity via get. name is the human-readable entity name used in error
// messages (e.g. "Assessment").
func fromParam[T any](ctx context.Context, param, name string, get func(context.Context, uuid.UUID) (*T, error)) (*T, string) {
	if param == "" {
		return nil, name + " ID is required"
	}

	id, err := util.ParseUUID(param)
	if err != nil {
		return nil, "Invalid " + strings.ToLower(name) + " ID"
	}

	out, err := get(ctx, id)
	if err != nil {
		return nil, "Invalid " + strings.ToLower(name) + " ID"
	}

	return out, ""
}

// userCustomerIDs returns the IDs of the customers the user is assigned to,
// or nil for admins (meaning "no customer filter").
func userCustomerIDs(user *model.User) []uuid.UUID {
	if user.Role == model.RoleAdmin {
		return nil
	}

	ids := make([]uuid.UUID, len(user.Customers))
	for i, customer := range user.Customers {
		ids[i] = customer.ID
	}
	return ids
}

// markOwnedAssessments sets IsOwned on the assessments owned by the user.
func markOwnedAssessments(user *model.User, assessments []model.Assessment) {
	owned := make(map[uuid.UUID]struct{}, len(user.Assessments))
	for _, ua := range user.Assessments {
		owned[ua.ID] = struct{}{}
	}
	for i := range assessments {
		if _, ok := owned[assessments[i].ID]; ok {
			assessments[i].IsOwned = true
		}
	}
}

func (d *Driver) gcFilesAsync() {
	go func() {
		if _, err := d.db.FileReference().GCFiles(context.Background()); err != nil {
			d.logger.Warn().Err(err).Msg("file gc failed")
		}
	}()
}

func (d *Driver) formDataReadFile(c *fiber.Ctx, fieldName string) (data []byte, filename string, err error) {
	file, err := c.FormFile(fieldName)
	if err != nil {
		return nil, "", err
	}

	data, err = d.readFile(file)
	if err != nil {
		return nil, "", err
	}

	return data, file.Filename, nil
}

func (d *Driver) formDataReadImage(c *fiber.Ctx, ctx context.Context, fieldName string) (data []byte, filename string, err error) {
	file, err := c.FormFile(fieldName)
	if err != nil {
		return nil, "", err
	}

	if file.Size == 0 {
		return nil, "", errors.New("Invalid file")
	}

	err = d.db.Setting().ValidateImageSize(ctx, file.Size)
	if err != nil {
		return nil, "", err
	}

	data, err = d.readFile(file)
	if err != nil {
		return nil, "", err
	}

	if !model.IsImageTypeAllowed(data) {
		return nil, "", store.ErrImageTypeNotAllowed
	}

	return data, file.Filename, nil
}

func (d *Driver) readFile(file *multipart.FileHeader) ([]byte, error) {
	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return data, nil
}
