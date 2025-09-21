package database

import (
	"fmt"
	"os"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"sister/internal/models"
)

var (
	dbInstance *gorm.DB
	dbOnce     sync.Once
)

func Init() *gorm.DB {
	dbOnce.Do(
		func() {
			dbHost := os.Getenv("DB_HOST")
			dbPort := os.Getenv("DB_PORT")
			dbName := os.Getenv("DB_NAME")
			dbUser := os.Getenv("DB_USER")
			dbPassword := os.Getenv("DB_PASSWORD")

			dsn := fmt.Sprintf(
				"host=%s user=%s dbname=%s port=%s sslmode=disable",
				dbHost, dbUser, dbName, dbPort,
			)
			if dbPassword != "" {
				dsn = fmt.Sprintf(
					"%s password=%s",
					dsn, dbPassword,
				)
			}

			db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
			if err != nil {
				panic("failed to connect to database: " + err.Error())
			}

			_ = db.AutoMigrate(
				&models.Message{},
			)

			dbInstance = db
		},
	)

	return dbInstance
}
