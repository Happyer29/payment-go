package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"payment-go/internal/config"
)

var db *gorm.DB = nil

func NewConnection(config *config.Config) (*gorm.DB, error) {
	dsn := GetDSNFromConfig(config)
	mysqlDb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db = mysqlDb

	return db, nil
}

func GetConnection() *gorm.DB {
	return db
}

func GetDSNFromConfig(config *config.Config) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.DbUser, config.DbPassword, config.DbHost, config.DbPort, config.DbName,
	)
}

func MakeMigrations(dst []interface{}) error {
	return db.AutoMigrate(dst...)
}
