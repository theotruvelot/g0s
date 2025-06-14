package database

import (
	"github.com/theotruvelot/g0s/internal/server/models"
	"github.com/theotruvelot/g0s/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(dsn string) (*gorm.DB, error) {
	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to the database", zap.Error(err))
		return nil, err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		logger.Error("Failed to get underlying sql.DB", zap.Error(err))
		return nil, err
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(3600) // 1 hour

	// Perform migration with proper error handling
	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		logger.Error("Failed to migrate models", zap.Error(err))
		return nil, err
	}

	logger.Info("Database connection established and models migrated successfully")
	return DB, nil
}

func GetDB() *gorm.DB {
	return DB
}

func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
