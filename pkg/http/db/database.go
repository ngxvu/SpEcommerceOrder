package db

import (
	"basesource/conf"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
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
	PgUser     string `json:"pgUser"`
	PgPassword string `json:"pgPassword"`
	PgHost     string `json:"pgHost"`
	PgPort     string `json:"pgPort"`
	PgDatabase string `json:"pgDatabase"`
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
		config.PgHost, config.PgPort, config.PgUser, config.PgPassword, config.PgDatabase)

	// Open database connection
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure read replicas
	dialector := postgres.New(postgres.Config{
		DSN: dsn,
	})

	if err = gormDB.Use(dbresolver.Register(dbresolver.Config{
		Replicas: []gorm.Dialector{dialector},
	})); err != nil {
		return nil, fmt.Errorf("failed to register db resolver: %w", err)
	}

	// Verify connection
	var result int
	if err = gormDB.Raw("SELECT 1").Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to verify database connection: %w", err)
	}

	log.Println("Database connection established")
	return gormDB, nil
}
