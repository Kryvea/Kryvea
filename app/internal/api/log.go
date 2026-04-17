package api

import (
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Kryvea/Kryvea/internal/log"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func (d *Driver) GetLog(c *fiber.Ctx) error {
	levels := c.Query("levels")

	allLevels := []zerolog.Level{
		zerolog.DebugLevel,
		zerolog.InfoLevel,
		zerolog.WarnLevel,
		zerolog.ErrorLevel,
		zerolog.FatalLevel,
		zerolog.PanicLevel,
	}

	// initialize all levels to false
	levelsMap := make(map[string]bool, len(allLevels))
	for _, lvl := range allLevels {
		levelsMap[lvl.String()] = false
	}

	// process user-specified levels
	i := 0
	for _, level := range strings.Split(levels, ",") {
		level = strings.ToLower(strings.TrimSpace(level))
		if parsedLevel, err := zerolog.ParseLevel(level); err == nil {
			if _, exists := levelsMap[parsedLevel.String()]; exists {
				levelsMap[parsedLevel.String()] = true
				i++
			}
		}
	}

	// if no valid levels specified, enable all
	if i == 0 {
		for _, lvl := range allLevels {
			levelsMap[lvl.String()] = true
		}
	}

	// parse pagination parameters
	page := c.Query("page", "1")
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	pageSize := c.Query("page_size", "0")
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeInt < 0 {
		pageSizeInt = 0
	}

	// open log file
	file, err := os.Open(log.GetLogPath())
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Failed to open log file",
		})
	}
	defer file.Close()

	results := []sonic.NoCopyRawMessage{}
	decoder := sonic.ConfigStd.NewDecoder(file)

	start := pageSizeInt * (pageInt - 1)
	end := pageSizeInt * pageInt

	line := 0
	for {
		var logEntry sonic.NoCopyRawMessage
		if err := decoder.Decode(&logEntry); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		// retrieve the log level from the JSON entry
		levelNode, err := sonic.Get(logEntry, "level")
		if err != nil {
			continue
		}

		levelStr, err := levelNode.String()
		if err != nil || !levelsMap[levelStr] {
			continue
		}

		// if log level is in the specified levels and within pagination range
		// append to results
		if levelsMap[levelStr] && (line >= start && (pageSizeInt == 0 || line < end)) {
			results = append(results, logEntry)
		}

		line++
	}

	totalPages := line
	if pageSizeInt > 0 {
		totalPages = (line + pageSizeInt - 1) / pageSizeInt
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"logs":        results,
		"page":        pageInt,
		"page_size":   pageSizeInt,
		"total_pages": totalPages,
	})
}
