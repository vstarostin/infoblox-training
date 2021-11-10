package service

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/vstarostin/infoblox-training-project-1/internal/model"
)

const (
	ErrUserAlreadyExist      = "user with phone %s already exists. Please write a correct one"
	ErrAddressBookIsEmpty    = "addressBook is empty"
	ErrNoSuchUserWithPhone   = "no such user with phone: %v"
	ErrNoSuchUserWithName    = "no such user with name: %v"
	DeleteUserMethodResponse = "%d user(s) was(were) deleted"
	ErrPhoneIsTaken          = "phone %v is already taken. Please write a correct one"
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
	LoadByName(name string) []model.User
	LoadByPhone(phone string) []model.User
	Store(user model.User) *gorm.DB
	Delete(name string) *gorm.DB
	Update(phone string, user model.User) *gorm.DB
}

func (abs *AddressBookService) AddUser(name, phone, address string) error {
	u := model.User{Name: name, Phone: phone, Address: address}
	result := abs.storage.Store(u)
	if result.Error != nil {
		return fmt.Errorf(ErrUserAlreadyExist, phone)
	}
	return nil
}

func (abs *AddressBookService) ListUsers() ([]model.User, error) {
	name := "%"
	users := abs.storage.LoadByName(name)
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
	users := abs.storage.LoadByName(name)
	if len(users) == 0 {
		return []model.User{}, fmt.Errorf(ErrNoSuchUserWithName, name)
	}
	return users, nil
}

func (abs *AddressBookService) DeleteUser(name string) (string, error) {
	if name == "" {
		name = "%"
	}
	name = strings.ReplaceAll(name, "*", "%")
	result := abs.storage.Delete(name)
	if result.RowsAffected == 0 {
		return "", fmt.Errorf(ErrNoSuchUserWithName, name)
	}
	return fmt.Sprintf(DeleteUserMethodResponse, result.RowsAffected), nil
}

func (abs *AddressBookService) UpdateUser(phone string, updatedUser model.User) error {
	existingUser := abs.storage.LoadByPhone(phone)
	if len(existingUser) == 0 {
		return fmt.Errorf(ErrNoSuchUserWithPhone, phone)
	}
	result := abs.storage.Update(phone, updatedUser)
	if result.Error != nil {
		return fmt.Errorf(ErrPhoneIsTaken, updatedUser.Phone)
	}
	return nil
}
