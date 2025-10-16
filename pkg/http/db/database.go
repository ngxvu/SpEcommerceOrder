package db

import (
	"basesource/conf"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"sync"
)

type DatabaseConfig struct {
	Read struct {
		Hostname   string
		Name       string
		Username   string
		Password   string
		Port       int
		Parameter  string
		DriverConn string
	}
}

// DatabaseCredentials represents standardized database credentials structure
type DatabaseCredentials struct {
	postgresUser     string `json:"postgresUser"`
	postgresPassword string `json:"postgresPassword"`
	postgresHost     string `json:"postgresHost"`
	postgresPort     string `json:"postgresPort"`
	postgresDatabase string `json:"postgresDatabase"`
}

var (
	dbInstance *gorm.DB
	once       sync.Once
)

// InitDatabase initializes and returns a singleton gorm database connection
func InitDatabase(config *conf.Config) (*gorm.DB, error) {
	var err error
	once.Do(func() {
		dbInstance, err = initializeDatabase(config)
	})
	return dbInstance, err
}

// initializeDatabase creates and configures the database connection
func initializeDatabase(config *conf.Config) (*gorm.DB, error) {
	// Create connection string using environment config
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.PostgresHost, config.PostgresPort, config.PostgresUser, config.PostgresPassword, config.PostgresDatabase)

	// Open database connection
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Verify connection
	var result int
	if err = gormDB.Raw("SELECT 1").Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to verify database connection: %w", err)
	}

	log.Println("Database connection established")
	return gormDB, nil
}
