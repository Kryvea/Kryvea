package util

import (
	"fmt"
	"io"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func ParseFormFile(c *fiber.Ctx, param string) ([]byte, error) {
	if param == "" {
		return nil, nil
	}

	fileHeader, err := c.FormFile(param)
	if err != nil {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func CreateImageReference(mime string, id uuid.UUID) string {
	extension := mimetype.Lookup(mime)
	if extension == nil {
		extension = mimetype.Lookup("image/png")
	}

	newFilename := fmt.Sprintf("%s%s", id.String(), extension.Extension())
	return newFilename
}
