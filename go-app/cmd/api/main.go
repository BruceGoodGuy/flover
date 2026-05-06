package main

import (
	"BruceGoodGuy/flover/pkg/cache"
	"BruceGoodGuy/flover/pkg/database"
	"BruceGoodGuy/flover/pkg/mail"
	"fmt"
	"os"

	"BruceGoodGuy/flover/internal/app"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"), os.Getenv("DB_SSL_MODE"),
		os.Getenv("DB_TIMEZONE"))
	db := database.Connect(dsn)

	rdb := cache.Connect()

	mb := new(mail.Mail).NewMail()

	application := app.NewAppContainer(db, rdb, mb)

	application.Routes()
}
