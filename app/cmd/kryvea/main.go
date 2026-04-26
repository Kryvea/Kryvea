package main

import (
	"github.com/Kryvea/Kryvea/internal/config"
	"github.com/Kryvea/Kryvea/internal/engine"
	"github.com/Kryvea/Kryvea/internal/i18n"
	"github.com/Kryvea/Kryvea/internal/log"
)

func main() {
	err := i18n.InitI18n(config.GetLocalesPath())
	if err != nil {
		return
	}

	levelWriter := log.NewLevelWriter(
		config.GetLogDirectory(),
		config.GetLogMaxSizeMB(),
		config.GetLogMaxBackups(),
		config.GetLogMaxAgeDays(),
		config.GetLogCompress(),
	)

	engine, err := engine.NewEngine(
		config.GetListeningAddr(),
		config.GetRootPath(),
		config.GetBodyLimitMB(),
		config.GetMongoURI(),
		config.GetAdminUser(),
		config.GetAdminPass(),
		levelWriter,
	)
	if err != nil {
		return
	}

	engine.Serve()
}
