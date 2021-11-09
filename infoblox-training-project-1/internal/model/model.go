package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name    string `gorm:"primaryKey"`
	Phone   string
	Address string
}
