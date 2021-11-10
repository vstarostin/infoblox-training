package service

import (
	"fmt"
	"strings"

	"github.com/vstarostin/infoblox-training-project-1/internal/model"
	"gorm.io/gorm"
)

const (
	ErrUserAlreadyExist      = "user with name %s already exists. Please choose different name"
	ErrAddressBookIsEmpty    = "addressBook is empty"
	ErrNoSuchUser            = "no such user with name: %v"
	DeleteUserMethodResponse = "%d user(s) was(were) deleted"
	ErrNameIsTaken           = "name %v is already taken. Please choose different name"
)

type AddressBookService struct {
	storage AddressBookStorage
}

func New(storage AddressBookStorage) *AddressBookService {
	return &AddressBookService{
		storage: storage,
	}
}

type AddressBookStorage interface {
	Load(name string) []model.User
	Store(user model.User)
	Delete(name string) *gorm.DB
	Update(user model.User)
}

func (abs *AddressBookService) AddUser(name, phone, address string) error {
	user := abs.storage.Load(name)
	if len(user) != 0 {
		return fmt.Errorf(ErrUserAlreadyExist, name)
	}
	u := model.User{Name: name, Phone: phone, Address: address}
	abs.storage.Store(u)
	return nil
}

func (abs *AddressBookService) ListUsers() ([]model.User, error) {
	name := "%"
	users := abs.storage.Load(name)
	if len(users) == 0 {
		return []model.User{}, fmt.Errorf(ErrAddressBookIsEmpty)
	}
	return users, nil
}

func (abs *AddressBookService) FindUser(name string) ([]model.User, error) {
	if name == "" || name == "*" {
		return abs.ListUsers()
	}
	name = strings.ReplaceAll(name, "*", "%")
	users := abs.storage.Load(name)
	if len(users) == 0 {
		return []model.User{}, fmt.Errorf(ErrNoSuchUser, name)
	}
	return users, nil
}

func (abs *AddressBookService) DeleteUser(name string) (string, error) {
	if name == "" || name == "*" {
		name = "%"
		result := abs.storage.Delete(name)
		if result.RowsAffected == 0 {
			return "", fmt.Errorf(ErrAddressBookIsEmpty)
		}
		return fmt.Sprintf(DeleteUserMethodResponse, result.RowsAffected), nil
	}
	name = strings.ReplaceAll(name, "*", "%")
	result := abs.storage.Delete(name)
	if result.RowsAffected == 0 {
		return "", fmt.Errorf(ErrNoSuchUser, name)
	}
	return fmt.Sprintf(DeleteUserMethodResponse, result.RowsAffected), nil
}

func (abs *AddressBookService) UpdateUser(name string, updatedUser model.User) error {
	user := abs.storage.Load(name)
	if len(user) == 0 {
		return fmt.Errorf(ErrNoSuchUser, name)
	}
	user = abs.storage.Load(updatedUser.Name)
	if len(user) != 0 {
		return fmt.Errorf(ErrNameIsTaken, name)
	}
	abs.storage.Update(updatedUser)
	return nil
}
