package log

import (
	"io"
	"os"
	"path/filepath"

	"github.com/Kryvea/Kryvea/internal/config"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logFileName = "kryvea.log"
)

type levelWriter struct {
	writer   io.Writer
	minLevel zerolog.Level
	maxLevel zerolog.Level
}

func (lw levelWriter) Write(p []byte) (n int, err error) {
	return lw.writer.Write(p)
}

func (lw levelWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	if level >= lw.minLevel && level <= lw.maxLevel {
		return lw.writer.Write(p)
	}
	return len(p), nil
}

func NewLevelWriter(logPath string, maxSizeMB, maxBackups, maxAgeDays int, compress bool) *zerolog.LevelWriter {
	logWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logPath, logFileName),
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays,
		Compress:   compress,
	}

	levelWriter := zerolog.MultiLevelWriter(
		levelWriter{writer: os.Stdout, minLevel: zerolog.DebugLevel, maxLevel: zerolog.PanicLevel},
		levelWriter{writer: logWriter, minLevel: zerolog.DebugLevel, maxLevel: zerolog.PanicLevel},
	)

	return &levelWriter
}

func GetLogPath() string {
	return filepath.Join(config.GetLogDirectory(), logFileName)
}
