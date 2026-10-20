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

// logPath is resolved by NewLevelWriter and read back by GetLogPath, whose
// callers do not have the log configuration at hand.
var logPath string

// NewLevelWriter returns a writer that duplicates every log line to
// stdout and to a size-rotated log file in cfg.Directory.
func NewLevelWriter(cfg config.Log) zerolog.LevelWriter {
	logPath = filepath.Join(cfg.Directory, logFileName)

	logWriter := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}

	return zerolog.MultiLevelWriter(os.Stdout, logWriter)
}

// GetLogPath returns the path of the current log file, resolved when
// NewLevelWriter was called.
func GetLogPath() string {
	return logPath
}
