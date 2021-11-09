package service

import (
	"fmt"
	"strings"

	"github.com/vstarostin/infoblox-training-project-1/internal/model"
	"gorm.io/gorm"
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
	// LoadAll() []model.User
	Store(user model.User)
	Delete(name string) *gorm.DB
	Update(user model.User) *gorm.DB
}

func (abs *AddressBookService) AddUser(name, phone, address string) error {
	user := abs.storage.Load(name)
	if len(user) != 0 {
		return fmt.Errorf("user with name %s already exists. Please choose different name", name)
	}
	u := model.User{Name: name, Phone: phone, Address: address}
	abs.storage.Store(u)
	return nil
}

func (abs *AddressBookService) ListUsers() ([]model.User, error) {
	name := "%"
	users := abs.storage.Load(name)
	if len(users) == 0 {
		return []model.User{}, fmt.Errorf("AddressBook is empty")
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
		return []model.User{}, fmt.Errorf("user is not found")
	}
	return users, nil
}

func (abs *AddressBookService) DeleteUser(name string) (string, error) {
	if name == "" || name == "*" {
		name = "%"
		result := abs.storage.Delete(name)
		if result.RowsAffected == 0 {
			return "", fmt.Errorf("AddressBook is empty")
		}
		return fmt.Sprintf("%d user(s) was(were) successfully deleted", result.RowsAffected), nil
	}
	name = strings.ReplaceAll(name, "*", "%")
	result := abs.storage.Delete(name)
	if result.RowsAffected == 0 {
		return "", fmt.Errorf("user does not exists")
	}
	return fmt.Sprintf("%d user(s) was(were) successfully deleted", result.RowsAffected), nil
}

func (abs *AddressBookService) UpdateUser(name string, updatedUser model.User) (model.User, error) {
	user := abs.storage.Load(name)
	if len(user) == 0 {
		return model.User{}, fmt.Errorf("user is not found")
	}
	//uncomment
	// user = abs.storage.Load(updatedUser.Name)
	// if len(user) != 0 {
	// 	return model.User{}, fmt.Errorf("name %v is already taken. Please choose different name.")
	// }
	result := abs.storage.Update(updatedUser)

	fmt.Printf("\n%d fields was updated\n\n", result.RowsAffected)

	return updatedUser, nil
}
