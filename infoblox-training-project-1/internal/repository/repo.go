package repository

import (
	"fmt"

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

func (s *Storage) LoadAll() []model.User {
	var users []model.User
	s.db.Find(&users)
	return users
}

func (s *Storage) DeleteAll() error {
	result := s.db.Exec(`DELETE FROM users`)
	if result.RowsAffected == 0 {
		return fmt.Errorf("AddressBook is empty")
	}
	return nil
}
