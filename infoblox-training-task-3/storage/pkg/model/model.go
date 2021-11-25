package model

import "github.com/jinzhu/gorm"

type ResponderMode struct {
	gorm.Model
	Mode bool `gorm:"default:true"`
}
