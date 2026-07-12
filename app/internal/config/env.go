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
	return getEnvInt(bodyLimitEnv, 1_000)
}

// GetPgDSN returns the PostgreSQL connection string.
func GetPgDSN() string {
	return getEnvConfig(pgDSNEnv, "postgres://kryvea:kryvea@localhost:5432/kryvea?sslmode=disable")
}

// GetPgMaxConns maps to sql.DB.SetMaxOpenConns. 0 means "default".
func GetPgMaxConns() int {
	v := getEnvInt(pgMaxConnsEnv, 0)
	if v <= 0 {
		return 0
	}
	return v
}

// GetPgMinConns maps to sql.DB.SetMaxIdleConns, i.e. despite its name
// (KRYVEA_PG_MIN_CONNS, kept for deployment compatibility) it sets the
// maximum number of idle connections. 0 means "default".
func GetPgMinConns() int {
	v := getEnvInt(pgMinConnsEnv, 0)
	if v < 0 {
		return 0
	}
	return v
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
	return getEnvInt(logMaxSizeMBEnv, 10)
}

func GetLogMaxBackups() int {
	return getEnvInt(logMaxBackupsEnv, 5)
}

func GetLogMaxAgeDays() int {
	return getEnvInt(logMaxAgeDaysEnv, 0)
}

func GetLogCompress() bool {
	return getEnvBool(logCompressEnv, true)
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

func getEnvInt(envName string, defaultValue int) int {
	value, err := strconv.Atoi(os.Getenv(envName))
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvBool(envName string, defaultValue bool) bool {
	value, err := strconv.ParseBool(os.Getenv(envName))
	if err != nil {
		return defaultValue
	}

	return value
}
