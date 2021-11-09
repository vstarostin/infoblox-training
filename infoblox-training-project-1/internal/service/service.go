package service

import (
	"fmt"

	"github.com/vstarostin/infoblox-training-project-1/internal/model"
)

type AddressBookService struct {
	repo AddressBookStorage
}

func New(repo AddressBookStorage) *AddressBookService {
	return &AddressBookService{
		repo: repo,
	}
}

type AddressBookStorage interface {
	LoadAll() []model.User
	Store(model.User)
	DeleteAll() error
}

func (abs *AddressBookService) AddUser(name, phone, address string) error {
	user := model.User{
		Name:    name,
		Phone:   phone,
		Address: address,
	}
	abs.repo.Store(user)
	return nil
}

func (abs *AddressBookService) ListUsers() ([]model.User, error) {
	users := abs.repo.LoadAll()
	if len(users) == 0 {
		return []model.User{}, fmt.Errorf("%v", "ErrAddressBookIsEmpty")
	}
	return users, nil
}

func (abs *AddressBookService) DeleteUser(name string) (string, error) {
	if name == "" || name == "*" {
		err := abs.repo.DeleteAll()
		if err != nil {
			return "", err
		}
		///
	}
	///
	return "successfully deleted", nil
}
