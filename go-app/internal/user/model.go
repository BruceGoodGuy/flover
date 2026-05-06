package user

import "time"

type User struct {
	Id        uint `gorm:"primaryKey"`
	FirstName string
	LastName  string
	Email     string `gorm:"unique"`
	Birthday  *time.Time
	Password  string
	Status    string `gorm:"default:pending"`
	Role      string `gorm:"default:human"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
