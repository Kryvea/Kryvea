package config

import (
	"os"
	"strconv"
	"time"
)

// Config is the complete runtime configuration. See Load for the environment
// variable and default value behind every field.
type Config struct {
	Addr        string // KRYVEA_ADDR
	RootPath    string // KRYVEA_ROOT_PATH
	BodyLimitMB int    // KRYVEA_BODY_LIMIT_MB
	LocalesPath string // KRYVEA_LOCALES_PATH

	DB    DB
	Admin Admin
	Log   Log
}

// DB configures the PostgreSQL connection and the local directory used to
// store binary file payloads (logo, template files, PoC images). The pool
// fields map to the database/sql setters of the same name; zero leaves the
// database/sql default in place.
type DB struct {
	DSN      string // KRYVEA_PG_DSN
	FilesDir string // KRYVEA_FILES_DIR

	MaxOpenConns    int           // KRYVEA_PG_MAX_CONNS
	MaxIdleConns    int           // KRYVEA_PG_MIN_CONNS, named for deployment compatibility
	ConnMaxLifetime time.Duration // KRYVEA_PG_MAX_CONN_LIFETIME
	ConnMaxIdleTime time.Duration // KRYVEA_PG_MAX_CONN_IDLE_TIME
}

// Admin holds the credentials of the account created on first startup.
type Admin struct {
	User string // KRYVEA_ADMIN_USER
	Pass string // KRYVEA_ADMIN_PASS
}

// Log configures the rotating log file written alongside stdout.
type Log struct {
	Directory  string // KRYVEA_LOG_DIRECTORY
	MaxSizeMB  int    // KRYVEA_LOG_MAX_SIZE_MB
	MaxBackups int    // KRYVEA_LOG_MAX_BACKUPS
	MaxAgeDays int    // KRYVEA_LOG_MAX_AGE_DAYS
	Compress   bool   // KRYVEA_LOG_COMPRESS
}

// Load reads the whole configuration from the environment. Every default is
// the second argument of its getEnv call below. A variable that is unset or
// malformed falls back to that default, so Load never fails.
func Load() Config {
	return Config{
		Addr:        getEnvConfig("KRYVEA_ADDR", "127.0.0.1:8000"),
		RootPath:    getEnvConfig("KRYVEA_ROOT_PATH", "/"),
		BodyLimitMB: getEnvInt("KRYVEA_BODY_LIMIT_MB", 1_000),
		LocalesPath: getEnvConfig("KRYVEA_LOCALES_PATH", "/etc/kryvea/locales"),

		DB: DB{
			DSN:             getEnvConfig("KRYVEA_PG_DSN", "postgres://kryvea:kryvea@localhost:5432/kryvea?sslmode=disable"),
			FilesDir:        getEnvConfig("KRYVEA_FILES_DIR", "/var/lib/kryvea/files"),
			MaxOpenConns:    getEnvInt("KRYVEA_PG_MAX_CONNS", 0),
			MaxIdleConns:    getEnvInt("KRYVEA_PG_MIN_CONNS", 0),
			ConnMaxLifetime: getEnvDuration("KRYVEA_PG_MAX_CONN_LIFETIME", 0),
			ConnMaxIdleTime: getEnvDuration("KRYVEA_PG_MAX_CONN_IDLE_TIME", 0),
		},

		Admin: Admin{
			User: getEnvConfig("KRYVEA_ADMIN_USER", "kryvea"),
			Pass: getEnvConfig("KRYVEA_ADMIN_PASS", "kryveapassword"),
		},

		Log: Log{
			Directory:  getEnvConfig("KRYVEA_LOG_DIRECTORY", "/var/log/kryvea/"),
			MaxSizeMB:  getEnvInt("KRYVEA_LOG_MAX_SIZE_MB", 10),
			MaxBackups: getEnvInt("KRYVEA_LOG_MAX_BACKUPS", 5),
			MaxAgeDays: getEnvInt("KRYVEA_LOG_MAX_AGE_DAYS", 0),
			Compress:   getEnvBool("KRYVEA_LOG_COMPRESS", true),
		},
	}
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

func getEnvDuration(envName string, defaultValue time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(envName))
	if err != nil {
		return defaultValue
	}

	return value
}
