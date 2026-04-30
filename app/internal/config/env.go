package config

import (
	"os"
	"strconv"
	"time"
)

const (
	addrEnv          = "KRYVEA_ADDR"
	rootPathEnv      = "KRYVEA_ROOT_PATH"
	bodyLimitEnv     = "KRYVEA_BODY_LIMIT_MB"
	pgDSNEnv         = "KRYVEA_PG_DSN"
	pgMaxConnsEnv    = "KRYVEA_PG_MAX_CONNS"
	pgMinConnsEnv    = "KRYVEA_PG_MIN_CONNS"
	pgMaxConnLifeEnv = "KRYVEA_PG_MAX_CONN_LIFETIME"
	pgMaxConnIdleEnv = "KRYVEA_PG_MAX_CONN_IDLE_TIME"
	filesDirEnv      = "KRYVEA_FILES_DIR"
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

func GetBodyLimitMB() int {
	defaultLimit := 1_000
	limit := getEnvConfig(bodyLimitEnv, "")

	bodyLimit, err := strconv.Atoi(limit)
	if err != nil {
		return defaultLimit
	}

	return bodyLimit
}

// GetPgDSN returns the PostgreSQL connection string.
func GetPgDSN() string {
	return getEnvConfig(pgDSNEnv, "postgres://kryvea:kryvea@localhost:5432/kryvea?sslmode=disable")
}

// GetPgMaxConns maps to sql.DB.SetMaxOpenConns. 0 means "default".
func GetPgMaxConns() int32 {
	v, err := strconv.Atoi(os.Getenv(pgMaxConnsEnv))
	if err != nil || v <= 0 {
		return 0
	}
	return int32(v)
}

// GetPgMinConns maps to sql.DB.SetMaxIdleConns. 0 means "default".
func GetPgMinConns() int32 {
	v, err := strconv.Atoi(os.Getenv(pgMinConnsEnv))
	if err != nil || v < 0 {
		return 0
	}
	return int32(v)
}

// GetPgMaxConnLifetime maps to sql.DB.SetConnMaxLifetime. 0 means "default".
func GetPgMaxConnLifetime() time.Duration {
	d, err := time.ParseDuration(os.Getenv(pgMaxConnLifeEnv))
	if err != nil || d <= 0 {
		return 0
	}
	return d
}

// GetPgMaxConnIdleTime maps to sql.DB.SetConnMaxIdleTime. 0 means "default".
func GetPgMaxConnIdleTime() time.Duration {
	d, err := time.ParseDuration(os.Getenv(pgMaxConnIdleEnv))
	if err != nil || d <= 0 {
		return 0
	}
	return d
}

// GetFilesDir returns the local directory used to store binary file payloads
// (logo, template files, PoC images).
func GetFilesDir() string {
	return getEnvConfig(filesDirEnv, "/var/lib/kryvea/files")
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
