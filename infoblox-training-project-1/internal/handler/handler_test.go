package handler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/vstarostin/infoblox-training-project-1/internal/handler"
	"github.com/vstarostin/infoblox-training-project-1/internal/mock"
	"github.com/vstarostin/infoblox-training-project-1/internal/model"
	"github.com/vstarostin/infoblox-training-project-1/internal/pb"
)

var (
	name, phone, address = "name", "phone", "address"
	err                  = errors.New("some error")
	user                 = &pb.User{UserName: name, Phone: phone, Address: address}
	users                = []*pb.User{user}
	modelUser            = model.User{Name: name, Phone: phone, Address: address}
	modelUsers           = []model.User{modelUser}
	emptyModelUsers      = []model.User{}
	responseOK           = "OK"
)

type handlerTestSuite struct {
	suite.Suite
	service *mock.AddressBookService
	handler *handler.AddressBook
}

func (suite *handlerTestSuite) SetupTest() {
	service := &mock.AddressBookService{}
	suite.service = service
	h := handler.New(service)
	suite.handler = h
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(handlerTestSuite))
}

func (suite *handlerTestSuite) TestHandlerAddUser() {
	tests := map[string]struct {
		serviceResponse  error
		expectedResponse *pb.AddUserResponse
		expectedErr      error
	}{
		"without_error": {
			serviceResponse:  nil,
			expectedResponse: &pb.AddUserResponse{Response: handler.AddUserMethodResponse},
			expectedErr:      nil,
		},
		"error": {
			serviceResponse:  err,
			expectedResponse: nil,
			expectedErr:      status.Error(codes.AlreadyExists, err.Error()),
		},
	}
	for testCase, test := range tests {
		suite.Run(testCase, func() {
			suite.service.On("AddUser", name, phone, address).Once().Return(test.serviceResponse)
			gotResponse, err := suite.handler.AddUser(context.Background(), &pb.AddUserRequest{NewUser: user})
			suite.Equal(test.expectedResponse, gotResponse)
			suite.Equal(test.expectedErr, err)
		})
	}
}

func (suite *handlerTestSuite) TestHandlerListUsers() {
	tests := map[string]struct {
		serviceUsersResponse []model.User
		serviceErrResponse   error
		expectedResponse     *pb.ListUsersResponse
		expectedErr          error
	}{
		"without_error": {
			serviceUsersResponse: modelUsers,
			serviceErrResponse:   nil,
			expectedResponse:     &pb.ListUsersResponse{Users: users},
			expectedErr:          nil,
		},
		"error": {
			serviceUsersResponse: emptyModelUsers,
			serviceErrResponse:   err,
			expectedResponse:     nil,
			expectedErr:          status.Error(codes.NotFound, err.Error()),
		},
	}
	for testCase, test := range tests {
		suite.Run(testCase, func() {
			suite.service.On("ListUsers").Once().Return(test.serviceUsersResponse, test.serviceErrResponse)
			gotResponse, err := suite.handler.ListUsers(context.Background(), &emptypb.Empty{})
			suite.Equal(test.expectedResponse, gotResponse)
			suite.Equal(test.expectedErr, err)
		})
	}
}

func (suite *handlerTestSuite) TestHandlerDeleteUser() {
	tests := map[string]struct {
		serviceResponse  string
		serviceErr       error
		expectedResponse *pb.DeleteUserResponse
		expectedErr      error
	}{
		"without_error": {
			serviceResponse:  responseOK,
			serviceErr:       nil,
			expectedResponse: &pb.DeleteUserResponse{Response: responseOK},
			expectedErr:      nil,
		},
		"error": {
			serviceResponse:  "",
			serviceErr:       err,
			expectedResponse: nil,
			expectedErr:      status.Error(codes.InvalidArgument, err.Error()),
		},
	}
	for testCase, test := range tests {
		suite.Run(testCase, func() {
			suite.service.On("DeleteUser", name).Once().Return(test.serviceResponse, test.serviceErr)
			gotResponse, err := suite.handler.DeleteUser(context.Background(), &pb.DeleteUserRequest{UserName: name})
			suite.Equal(test.expectedResponse, gotResponse)
			suite.Equal(test.expectedErr, err)
		})
	}
}

func (suite *handlerTestSuite) TestHandlerFindUser() {
	tests := map[string]struct {
		serviceResponse  []model.User
		serviceErr       error
		expectedResponse *pb.FindUserResponse
		expectedErr      error
	}{
		"without_error": {
			serviceResponse:  modelUsers,
			serviceErr:       nil,
			expectedResponse: &pb.FindUserResponse{Users: users},
			expectedErr:      nil,
		},
		"error": {
			serviceResponse:  emptyModelUsers,
			serviceErr:       err,
			expectedResponse: nil,
			expectedErr:      status.Error(codes.InvalidArgument, err.Error()),
		},
	}
	for testCase, test := range tests {
		suite.Run(testCase, func() {
			suite.service.On("FindUser", name, "", "").Once().Return(test.serviceResponse, test.serviceErr)
			gotResponse, err := suite.handler.FindUser(context.Background(), &pb.FindUserRequest{Name: name})
			suite.Equal(test.expectedResponse, gotResponse)
			suite.Equal(test.expectedErr, err)
		})
	}
}

func (suite *handlerTestSuite) TestHandlerUpdateUser() {
	tests := map[string]struct {
		serviceResponse  error
		expectedResponse *pb.UpdateUserResponse
		expectedErr      error
	}{
		"without_error": {
			serviceResponse:  nil,
			expectedResponse: &pb.UpdateUserResponse{Response: handler.UpdateUserMethodResponse, UpdatedUser: user},
			expectedErr:      nil,
		},
		"error": {
			serviceResponse:  err,
			expectedResponse: nil,
			expectedErr:      status.Error(codes.InvalidArgument, err.Error()),
		},
	}
	for testCase, test := range tests {
		suite.Run(testCase, func() {
			suite.service.On("UpdateUser", phone, modelUser).Once().Return(test.serviceResponse)
			gotResponse, err := suite.handler.UpdateUser(context.Background(), &pb.UpdateUserRequest{Phone: phone, UpdatedUser: user})
			suite.Equal(test.expectedResponse, gotResponse)
			suite.Equal(test.expectedErr, err)
		})
	}
}
