package api

import (
	"context"
	"errors"
	"io"
	"mime/multipart"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/gofiber/fiber/v2"
)

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
