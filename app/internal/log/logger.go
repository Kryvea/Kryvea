package log

import (
	"os"
	"path/filepath"

	"github.com/Kryvea/Kryvea/internal/config"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logFileName = "kryvea.log"
)

// NewLevelWriter returns a writer that duplicates every log line to
// stdout and to a size-rotated log file in logDir.
func NewLevelWriter(logDir string, maxSizeMB, maxBackups, maxAgeDays int, compress bool) zerolog.LevelWriter {
	logWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, logFileName),
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays,
		Compress:   compress,
	}

	return zerolog.MultiLevelWriter(os.Stdout, logWriter)
}

// GetLogPath returns the path of the current log file. It reads the log
// directory from config because its callers do not have it at hand.
func GetLogPath() string {
	return filepath.Join(config.GetLogDirectory(), logFileName)
}
