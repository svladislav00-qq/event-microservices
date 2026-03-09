package postgres

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// func MigrateAccounts(db *gorm.DB) error {
// 	return db.AutoMigrate(models...)
// }
