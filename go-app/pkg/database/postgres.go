package database

import (
	"BruceGoodGuy/flover/internal/test"
	"BruceGoodGuy/flover/internal/user"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("failed to connect database")
		panic(err)
	}

	db.AutoMigrate(&test.Test{}, &user.User{})

	return db
}
