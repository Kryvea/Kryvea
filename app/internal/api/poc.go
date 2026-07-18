package api

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/safe"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
)

var hexColorRegex = regexp.MustCompile(`^#?[a-fA-F0-9]{6}$`)

type pocData struct {
	Index              int                     `json:"index"`
	Type               string                  `json:"type"`
	Description        string                  `json:"description"`
	URI                string                  `json:"uri"`
	Request            string                  `json:"request"`
	RequestHighlights  []model.HighlightedText `json:"request_highlights"`
	Response           string                  `json:"response"`
	ResponseHighlights []model.HighlightedText `json:"response_highlights"`
	ImageReference     string                  `json:"image_reference"`
	ImageCaption       string                  `json:"image_caption"`
	TextLanguage       string                  `json:"text_language"`
	TextData           string                  `json:"text_data"`
	TextHighlights     []model.HighlightedText `json:"text_highlights"`
	StartingLineNumber int                     `json:"starting_line_number"`
}

func (d *Driver) UpsertPocs(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	vulnerability, errStr := d.vulnerabilityFromParam(c.UserContext(), c.Params("vulnerability"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	if !user.CanAccessCustomer(vulnerability.Customer.ID) {
		return jsonError(c, fiber.StatusForbidden, "Forbidden")
	}

	pocsData := []pocData{}
	pocsStr := c.FormValue("pocs")
	err := sonic.Unmarshal([]byte(pocsStr), &pocsData)
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	for i := range pocsData {
		errStr = d.validatePocData(&pocsData[i])
		if errStr != "" {
			return jsonError(c, fiber.StatusBadRequest, errStr)
		}
	}

	pocs := make([]model.PocItem, len(pocsData))
	safePocs := safe.New(pocs)

	errorChan := make(chan string, len(pocsData))

	wg := sync.WaitGroup{}
	// read image data in parallel; images are persisted later in the transaction
	for i, data := range pocsData {
		wg.Add(1)
		go func(i int, data pocData) {
			defer wg.Done()
			pocImageFilename := ""
			imageData := []byte{}
			if data.Type == model.PocTypeImage && data.ImageReference != "" {
				imageData, pocImageFilename, err = d.formDataReadImage(c, c.UserContext(), data.ImageReference)
				if err != nil {
					c.Status(fiber.StatusBadRequest)

					switch err {
					case store.ErrFileSizeTooLarge:
						errorChan <- fmt.Sprintf("PoC %d: Image file size is too large", i)
						return
					case store.ErrImageTypeNotAllowed:
						errorChan <- fmt.Sprintf("PoC %d: Image type is not allowed", i)
						return
					}

					errorChan <- fmt.Sprintf("PoC %d: Cannot read image data", i)
					return
				}
			}
			safePocs.Set(i, model.PocItem{
				Index:              data.Index,
				Type:               data.Type,
				Description:        data.Description,
				URI:                data.URI,
				Request:            data.Request,
				RequestHighlights:  data.RequestHighlights,
				Response:           data.Response,
				ResponseHighlights: data.ResponseHighlights,
				ImageData:          imageData,
				ImageFilename:      pocImageFilename,
				ImageCaption:       data.ImageCaption,
				TextLanguage:       data.TextLanguage,
				TextData:           data.TextData,
				TextHighlights:     data.TextHighlights,
				StartingLineNumber: data.StartingLineNumber,
			})
		}(i, data)
	}

	wg.Wait()
	close(errorChan)

	// Collect all errors
	var errs []string
	for err := range errorChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error":  "Failed to process pocs",
			"errors": errs,
		})
	}

	_, err = d.db.RunInTx(c.UserContext(), func(ctx context.Context) (any, error) {
		pocs := safePocs.GetAll()
		for i := range pocs {
			// only image pocs carry image data; skip the file insert for the rest
			if len(pocs[i].ImageData) == 0 {
				continue
			}

			imageID, mime, err := d.db.FileReference().Insert(ctx, pocs[i].ImageData)
			if err != nil {
				return nil, fmt.Errorf("PoC %d: Cannot upload image", i)
			}

			pocs[i].ImageID = imageID
			pocs[i].ImageMimeType = mime
		}

		pocUpsert := &model.Poc{
			VulnerabilityID: vulnerability.ID,
			Pocs:            pocs,
		}

		err = d.db.Poc().Upsert(ctx, pocUpsert)
		if err != nil {
			return nil, errors.New("Failed to update PoC")
		}

		return nil, nil
	})
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, err.Error())
	}

	d.gcFilesAsync()
	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "PoCs updated",
	})
}

func (d *Driver) GetPocsByVulnerability(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	vulnerability, errStr := d.vulnerabilityFromParam(c.UserContext(), c.Params("vulnerability"))
	if errStr != "" {
		return jsonError(c, fiber.StatusBadRequest, errStr)
	}

	if !user.CanAccessCustomer(vulnerability.Customer.ID) {
		return jsonError(c, fiber.StatusForbidden, "Forbidden")
	}

	poc, err := d.db.Poc().GetByVulnerabilityID(c.UserContext(), vulnerability.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.Status(fiber.StatusOK)
			return c.JSON([]model.PocItem{})
		}
		return jsonError(c, fiber.StatusInternalServerError, "Cannot get PoCs")
	}

	c.Status(fiber.StatusOK)
	return c.JSON(poc.Pocs)
}

func (d *Driver) validatePocData(data *pocData) string {
	if !model.IsValidPocType(data.Type) {
		return "Invalid PoC type"
	}

	for i, highlight := range data.RequestHighlights {
		if highlight.Color != "" && !hexColorRegex.MatchString(highlight.Color) {
			return fmt.Sprintf("Invalid color format for request highlight %d: %s", i, highlight.Color)
		}
	}
	for i, highlight := range data.ResponseHighlights {
		if highlight.Color != "" && !hexColorRegex.MatchString(highlight.Color) {
			return fmt.Sprintf("Invalid color format for response highlight %d: %s", i, highlight.Color)
		}
	}
	for i, highlight := range data.TextHighlights {
		if highlight.Color != "" && !hexColorRegex.MatchString(highlight.Color) {
			return fmt.Sprintf("Invalid color format for text highlight %d: %s", i, highlight.Color)
		}
	}

	switch data.Type {
	case model.PocTypeText:
		if strings.TrimSpace(data.TextData) == "" {
			return "Text data cannot be empty"
		}
	case model.PocTypeRequest:
		if strings.TrimSpace(data.Request) == "" && strings.TrimSpace(data.Response) == "" {
			return "Request and Response cannot be both empty"
		}
	case model.PocTypeImage:
		if strings.TrimSpace(data.ImageReference) == "" {
			return "Image reference cannot be empty"
		}
	default:
		return "Invalid PoC type"
	}

	data.Description = strings.Trim(data.Description, trimCutset)
	data.Request = strings.Trim(data.Request, trimCutset)
	data.Response = strings.Trim(data.Response, trimCutset)
	data.TextData = strings.Trim(data.TextData, trimCutset)

	return ""
}
