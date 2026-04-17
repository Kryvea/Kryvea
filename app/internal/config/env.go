package config

import (
	"os"
	"strconv"
)

const (
	addrEnv          = "KRYVEA_ADDR"
	rootPathEnv      = "KRYVEA_ROOT_PATH"
	mongoURIEnv      = "KRYVEA_MONGO_URI"
	adminUserEnv     = "KRYVEA_ADMIN_USER"
	adminPassEnv     = "KRYVEA_ADMIN_PASS"
	logDirectoryEnv  = "KRYVEA_LOG_DIRECTORY"
	logMaxSizeMBEnv  = "KRYVEA_LOG_MAX_SIZE_MB"
	logMaxBackupsEnv = "KRYVEA_LOG_MAX_BACKUPS"
	logMaxAgeDaysEnv = "KRYVEA_LOG_MAX_AGE_DAYS"
	logCompressEnv   = "KRYVEA_LOG_COMPRESS"
	localesPathEnv   = "KRYVEA_LOCALES_PATH"
)

func GetListeningAddr() string {
	return getEnvConfig(addrEnv, "127.0.0.1:8000")
}

func GetRootPath() string {
	return getEnvConfig(rootPathEnv, "/")
}

func GetMongoURI() string {
	return getEnvConfig(mongoURIEnv, "mongodb://user:password@host:27017")
}

func GetAdminUser() string {
	return getEnvConfig(adminUserEnv, "kryvea")
}

func GetAdminPass() string {
	return getEnvConfig(adminPassEnv, "kryveapassword")
}

func GetLogDirectory() string {
	return getEnvConfig(logDirectoryEnv, "/var/log/kryvea/")
}

func GetLogMaxSizeMB() int {
	defaultSize := 10
	size := getEnvConfig(logMaxSizeMBEnv, "")

	maxSize, err := strconv.Atoi(size)
	if err != nil {
		return defaultSize
	}

	return maxSize
}

func GetLogMaxBackups() int {
	defaultBackups := 5
	backups := os.Getenv(logMaxBackupsEnv)

	maxBackups, err := strconv.Atoi(backups)
	if err != nil {
		return defaultBackups
	}

	return maxBackups
}

func GetLogMaxAgeDays() int {
	defaultAge := 0
	age := os.Getenv(logMaxAgeDaysEnv)

	maxAge, err := strconv.Atoi(age)
	if err != nil {
		return defaultAge
	}

	return maxAge
}

func GetLogCompress() bool {
	defaultCompress := true
	compress := os.Getenv(logCompressEnv)

	compressBool, err := strconv.ParseBool(compress)
	if err != nil {
		return defaultCompress
	}

	return compressBool
}

func GetLocalesPath() string {
	return getEnvConfig(localesPathEnv, "/etc/kryvea/locales")
}

func getEnvConfig(envName, defaultValue string) string {
	value := os.Getenv(envName)
	if value != "" {
		return value
	}

	return defaultValue
}
