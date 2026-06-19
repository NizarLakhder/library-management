// Package database centralizes everything related to the PostgreSQL connection:
// building the DSN, validating user input, opening the GORM handle and closing
// it cleanly. Keeping this logic out of the UI makes the connection string
// (especially DSN escaping) unit-testable without a running database.
package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds the PostgreSQL connection parameters entered in the login form.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// DSN returns the libpq key=value connection string. The key=value format
// (rather than a URL) is used on purpose: it handles special characters in
// passwords (@, :, /, %) safely without URL-encoding.
func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=America/Toronto",
		c.Host, c.Port, c.User, c.Password, c.DBName,
	)
}

// Validate checks that every required field is present. The password is
// optional (a database account may have none).
func (c Config) Validate() error {
	if c.Host == "" || c.Port == "" || c.User == "" || c.DBName == "" {
		return errors.New("tous les champs (sauf mot de passe) sont requis")
	}
	return nil
}

// Connect validates the config, opens a GORM connection and pings the database
// to confirm it is actually reachable. A non-nil error means the returned *gorm.DB
// must not be used.
func Connect(c Config) (*gorm.DB, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(c.DSN()), &gorm.Config{Logger: gormLogger})
	if err != nil {
		return nil, fmt.Errorf("connexion GORM: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("obtention sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping DB: %w", err)
	}

	return db, nil
}

// Close releases the underlying connection pool. Safe to call with a nil handle.
func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
