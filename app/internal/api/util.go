package api

import (
	"context"
	"errors"
	"io"
	"mime/multipart"

	"github.com/Kryvea/Kryvea/internal/mongo"
	"github.com/gofiber/fiber/v2"
)

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

	err = d.mongo.Setting().ValidateImageSize(ctx, file.Size)
	if err != nil {
		return nil, "", err
	}

	data, err = d.readFile(file)
	if err != nil {
		return nil, "", err
	}

	if !mongo.IsImageTypeAllowed(data) {
		return nil, "", mongo.ErrImageTypeNotAllowed
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
