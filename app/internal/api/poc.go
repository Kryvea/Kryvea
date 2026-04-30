package api

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/poc"
	"github.com/Kryvea/Kryvea/internal/safe"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

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

	// parse vulnerability param
	vulnerability, errStr := d.vulnerabilityFromParam(c.UserContext(), c.Params("vulnerability"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// get assessment from database
	assessment, err := d.db.Assessment().GetByID(c.UserContext(), vulnerability.Assessment.ID)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Invalid vulnerability",
		})
	}

	// check if user can access the customer
	if !user.CanAccessCustomer(assessment.Customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	// parse request body
	pocsData := []pocData{}
	pocsStr := c.FormValue("pocs")
	err = sonic.Unmarshal([]byte(pocsStr), &pocsData)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// validate data
	for i := range pocsData {
		errStr = d.validatePocData(&pocsData[i])
		if errStr != "" {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": errStr,
			})
		}
	}

	pocs := make([]model.PocItem, len(pocsData))
	safePocs := safe.New(pocs)

	errorChan := make(chan string, len(pocsData))

	wg := sync.WaitGroup{}
	// parse image data and insert it into the database
	for i, data := range pocsData {
		wg.Add(1)
		go func(i int, data pocData) {
			defer wg.Done()
			imageID := uuid.UUID{}
			pocImageFilename := ""
			imageData := []byte{}
			if data.Type == poc.PocTypeImage && data.ImageReference != "" {
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
				ImageID:            imageID,
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

		// update poc in the database
		err = d.db.Poc().Upsert(ctx, pocUpsert)
		if err != nil {
			return nil, errors.New("Failed to update PoC")
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
		"message": "PoCs updated",
	})
}

func (d *Driver) GetPocsByVulnerability(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	// parse vulnerability param
	vulnerability, errStr := d.vulnerabilityFromParam(c.UserContext(), c.Params("vulnerability"))
	if errStr != "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": errStr,
		})
	}

	// get assessment from database
	assessment, err := d.db.Assessment().GetByID(c.UserContext(), vulnerability.Assessment.ID)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Invalid vulnerability",
		})
	}

	if !user.CanAccessCustomer(assessment.Customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	// parse vulnerability param
	poc, err := d.db.Poc().GetByVulnerabilityID(c.UserContext(), vulnerability.ID)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot get PoCs",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(poc.Pocs)
}

func (d *Driver) validatePocData(data *pocData) string {
	if !poc.IsValidType(data.Type) {
		return "Invalid PoC type"
	}

	hexColorRegex := regexp.MustCompile(`^#?[a-fA-F0-9]{6}$`)
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
	case poc.PocTypeText:
		if strings.TrimSpace(data.TextData) == "" {
			return "Text data cannot be empty"
		}
	case poc.PocTypeRequest:
		if strings.TrimSpace(data.Request) == "" && strings.TrimSpace(data.Response) == "" {
			return "Request and Response cannot be both empty"
		}
	case poc.PocTypeImage:
		if strings.TrimSpace(data.ImageReference) == "" {
			return "Image reference cannot be empty"
		}
	default:
		return "Invalid PoC type"
	}

	data.Description = strings.Trim(data.Description, "\r\n ")
	data.Request = strings.Trim(data.Request, "\r\n ")
	data.Response = strings.Trim(data.Response, "\r\n ")
	data.TextData = strings.Trim(data.TextData, "\r\n ")

	return ""
}
