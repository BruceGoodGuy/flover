package test

import "time"

type Test struct {
	Id        uint `gorm:"primaryKey"`
	Name      string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
