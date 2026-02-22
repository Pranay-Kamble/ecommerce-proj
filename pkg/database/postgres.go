package database

import (
	"ecommerce/pkg/logger"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Postgres struct {
	DB *gorm.DB
}

func NewPostgres(dsn string) (*Postgres, error) {
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return nil, err
	}

	return &Postgres{DB: db}, nil
}

func (p *Postgres) Connect(dsn string) error {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("database (postgres): failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("database (postgres): failed to get underlying database object: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	p.DB = db

	logger.Info("Connected to PostgreSQL")
	return nil
}

func (p *Postgres) Close() error {
	if p.DB == nil {
		return nil
	}

	sqlDb, err := p.DB.DB()
	if err != nil {
		return fmt.Errorf("database (postgres): failed to get underlying database object: %w", err)
	}

	err = sqlDb.Close()
	if err != nil {
		return fmt.Errorf("database (postgres): failed to close database: %w", err)
	}
	return nil
}
