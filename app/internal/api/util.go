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

// formDataReadImage reads a single uploaded image, resolving the configured
// size limit itself. Callers reading several images in a loop should fetch the
// limit once and use formDataReadImageMax instead.
func (d *Driver) formDataReadImage(c *fiber.Ctx, ctx context.Context, fieldName string) (data []byte, filename string, err error) {
	maxImageSize, err := d.maxImageSize(ctx)
	if err != nil {
		return nil, "", err
	}

	return d.formDataReadImageMax(c, fieldName, maxImageSize)
}

// formDataReadImageMax is formDataReadImage with the size limit already
// resolved, so a caller reading many images issues one settings query in total
// rather than one per image.
func (d *Driver) formDataReadImageMax(c *fiber.Ctx, fieldName string, maxImageSize int64) (data []byte, filename string, err error) {
	file, err := c.FormFile(fieldName)
	if err != nil {
		return nil, "", err
	}

	if file.Size == 0 {
		return nil, "", errors.New("Invalid file")
	}

	if file.Size > maxImageSize {
		return nil, "", store.ErrFileSizeTooLarge
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

// maxImageSize returns the configured upload limit for a single image.
func (d *Driver) maxImageSize(ctx context.Context) (int64, error) {
	setting, err := d.db.Setting().Get(ctx)
	if err != nil {
		return 0, err
	}

	return setting.MaxImageSize, nil
}

// readFile copies the whole uploaded part into memory. The buffer is sized
// upfront from file.Size: io.ReadAll would instead grow from 512 bytes,
// allocating roughly five times the payload in transient garbage.
func (d *Driver) readFile(file *multipart.FileHeader) ([]byte, error) {
	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data := make([]byte, file.Size)
	if _, err := io.ReadFull(f, data); err != nil {
		return nil, err
	}

	return data, nil
}
