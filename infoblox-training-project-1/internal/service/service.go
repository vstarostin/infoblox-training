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
	ErrUserDoesNotExist      = "user does not exist"
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
	Load(user model.User) []model.User
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
	name, phone, address := "%", "%", "%"
	user := model.User{Name: name, Phone: phone, Address: address}
	users := abs.storage.Load(user)
	if len(users) == 0 {
		return []model.User{}, fmt.Errorf(ErrAddressBookIsEmpty)
	}
	return users, nil
}

func (abs *AddressBookService) FindUser(name, phone, address string) ([]model.User, error) {
	if name == "" {
		name = "%"
	}
	if phone == "" {
		phone = "%"
	}
	if address == "" {
		address = "%"
	}
	name = strings.ReplaceAll(name, "*", "%")
	phone = strings.ReplaceAll(phone, "*", "%")
	address = strings.ReplaceAll(address, "*", "%")

	user := model.User{Name: name, Phone: phone, Address: address}
	users := abs.storage.Load(user)
	if len(users) == 0 {
		return []model.User{}, fmt.Errorf(ErrUserDoesNotExist)
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
	user := model.User{Name: "%", Address: "%", Phone: phone}
	existingUser := abs.storage.Load(user)
	if len(existingUser) == 0 {
		return fmt.Errorf(ErrUserDoesNotExist)
	}
	result := abs.storage.Update(phone, updatedUser)
	if result.Error != nil {
		return fmt.Errorf(ErrPhoneIsTaken, updatedUser.Phone)
	}
	return nil
}
