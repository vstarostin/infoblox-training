package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name    string
	Phone   string `gorm:"unique"`
	Address string
}
