package service_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/vstarostin/infoblox-training-project-1/internal/mock"
	"github.com/vstarostin/infoblox-training-project-1/internal/model"
	"github.com/vstarostin/infoblox-training-project-1/internal/service"
)

var (
	name, phone, address = "name", "phone", "address"
	user                 = model.User{Name: name, Phone: phone, Address: address}
	users                = []model.User{user}
	emptyUsers           = []model.User{}
)

type serviceTestSuite struct {
	suite.Suite
	service *service.AddressBookService
	storage *mock.AddressBookStorage
}

func (suite *serviceTestSuite) SetupTest() {
	storage := &mock.AddressBookStorage{}
	s := service.New(storage)
	suite.storage = storage
	suite.service = s
}

func TestService(t *testing.T) {
	suite.Run(t, new(serviceTestSuite))
}

func (suite *serviceTestSuite) TestServiceAddUser() {
	tests := map[string]struct {
		storageResponse *gorm.DB
		expectedResult  error
	}{
		"without_error": {
			storageResponse: &gorm.DB{},
			expectedResult:  nil,
		},
		"error": {
			storageResponse: &gorm.DB{Error: errors.New("some error")},
			expectedResult:  fmt.Errorf(service.ErrUserAlreadyExist, phone),
		},
	}

	for name, test := range tests {
		suite.Run(name, func() {
			suite.storage.On("Store", user).Once().Return(test.storageResponse)
			gotResult := suite.service.AddUser(user.Name, user.Phone, user.Address)
			suite.Equal(test.expectedResult, gotResult)
		})
	}
}

func (suite *serviceTestSuite) TestServiceListUsers() {
	name, address, phone := "%", "%", "%"
	user := model.User{Name: name, Phone: phone, Address: address}
	tests := map[string]struct {
		storageResponse, expectedResult []model.User
	}{
		"without_error": {
			storageResponse: users,
			expectedResult:  users,
		},
		"error": {
			storageResponse: emptyUsers,
			expectedResult:  emptyUsers,
		},
	}

	for caseName, test := range tests {
		suite.Run(caseName, func() {
			suite.storage.On("Load", user).Once().Return(test.storageResponse)
			gotResult, _ := suite.service.ListUsers()
			suite.Equal(test.expectedResult, gotResult)
		})
	}
}

func (suite *serviceTestSuite) TestServiceFindUser() {
	tests := map[string]struct {
		users, storageResponse, expectedResult []model.User
		expectedErr                            error
	}{
		"without_error": {
			users:           users,
			storageResponse: users,
			expectedResult:  users,
			expectedErr:     nil,
		},
		"error": {
			users:           emptyUsers,
			storageResponse: emptyUsers,
			expectedResult:  emptyUsers,
			expectedErr:     fmt.Errorf(service.ErrUserDoesNotExist),
		},
	}
	for caseName, test := range tests {
		suite.Run(caseName, func() {
			suite.storage.On("Load", user).Once().Return(test.users)
			gotResult, err := suite.service.FindUser(name, phone, address)
			suite.Equal(test.expectedResult, gotResult)
			suite.Equal(test.expectedErr, err)
		})
	}
}

func (suite *serviceTestSuite) TestServiceDeleteUser() {
	tests := map[string]struct {
		storageResponse *gorm.DB
		expectedResult  string
		expectedErr     error
	}{
		"without_error": {
			storageResponse: &gorm.DB{RowsAffected: 1},
			expectedResult:  fmt.Sprintf(service.DeleteUserMethodResponse, 1),
			expectedErr:     nil,
		},
		"error": {
			storageResponse: &gorm.DB{RowsAffected: 0},
			expectedResult:  "",
			expectedErr:     fmt.Errorf(service.ErrNoSuchUserWithName, name),
		},
	}
	for caseName, test := range tests {
		suite.Run(caseName, func() {
			suite.storage.On("Delete", name).Once().Return(test.storageResponse)
			gotResult, err := suite.service.DeleteUser(name)
			suite.Equal(test.expectedResult, gotResult)
			suite.Equal(test.expectedErr, err)
		})
	}
}

func (suite *serviceTestSuite) TestServiceUpdateUser() {
	u := model.User{Name: "%", Phone: phone, Address: "%"}
	tests := map[string]struct {
		users           []model.User
		storageResponse *gorm.DB
		expectedErr     error
	}{
		"without_error": {
			users:           users,
			storageResponse: &gorm.DB{},
			expectedErr:     nil,
		},
		"error_1": {
			users:           users,
			storageResponse: &gorm.DB{Error: fmt.Errorf("some error")},
			expectedErr:     fmt.Errorf(service.ErrPhoneIsTaken, phone),
		},
	}
	for caseName, test := range tests {
		suite.Run(caseName, func() {
			suite.storage.On("Load", u).Once().Return(test.users)
			suite.storage.On("Update", phone, user).Once().Return(test.storageResponse)
			err := suite.service.UpdateUser(u.Phone, user)
			suite.Equal(test.expectedErr, err)
		})
	}
}

func (suite *serviceTestSuite) TestServiceUpdateUserError() {
	u := model.User{Name: "%", Phone: phone, Address: "%"}
	expectedErr := fmt.Errorf(service.ErrUserDoesNotExist)
	suite.storage.On("Load", u).Once().Return(emptyUsers)
	err := suite.service.UpdateUser(u.Phone, user)
	suite.Equal(expectedErr, err)
}
