package database

import (
	"fmt"
	"log"
	"os"

	"securityrota-api/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "rota")
	password := getEnv("DB_PASSWORD", "rotapass")
	dbname := getEnv("DB_NAME", "securityrota")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")

	// Auto migrate models
	err = DB.AutoMigrate(&models.Officer{}, &models.Shift{}, &models.WeekRotation{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
