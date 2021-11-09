package repository

import (
	"gorm.io/gorm"

	"github.com/vstarostin/infoblox-training-project-1/internal/model"
)

type Storage struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Store(user model.User) {
	s.db.Select("name", "phone", "address").Create(&user)
}

func (s *Storage) Load(name string) []model.User {
	user := []model.User{}
	s.db.Where("name LIKE ?", name).Find(&user)
	return user
}

func (s *Storage) Delete(name string) *gorm.DB {
	return s.db.Exec("DELETE FROM users WHERE name LIKE ?", name)
}

func (s *Storage) Update(updaterUser model.User) *gorm.DB {
	return s.db.Exec("UPDATE users SET name=?, phone=?, address=? ", updaterUser.Name, updaterUser.Phone, updaterUser.Address)
}
