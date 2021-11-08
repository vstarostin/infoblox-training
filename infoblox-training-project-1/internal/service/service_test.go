package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	"github.com/vstarostin/infoblox-training-project-1/internal/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	UserName   = "name"
	Phone      = "phone"
	Address    = "address"
	NewName    = "newname"
	NewPhone   = "newphone"
	NewAddress = "newaddress"
	User       = &pb.User{UserName: UserName, Phone: Phone, Address: Address}
)

func TestAddUser(t *testing.T) {
	ctx := context.Background()
	addressBook := New()
	user := User

	expectedResponse_without_error := &pb.AddUserResponse{Response: AddUserMethodResponse}
	addUserRequest := &pb.AddUserRequest{
		NewUser: user,
	}

	tests := map[string]struct {
		expectedResponse *pb.AddUserResponse
		expectedErr      error
	}{
		"without_error": {
			expectedResponse: expectedResponse_without_error,
			expectedErr:      nil,
		},
		"error": {
			expectedResponse: nil,
			expectedErr:      status.Errorf(codes.AlreadyExists, ErrUserAlreadyExist, addUserRequest.GetNewUser().GetUserName())},
	}

	for name, test := range tests {
		switch name {
		case "error":
			addressBook.data[user.GetUserName()] = user
		default:
			delete(addressBook.data, user.GetUserName())
		}
		t.Run(name, func(t *testing.T) {
			gotResponse, err := addressBook.AddUser(ctx, addUserRequest)
			assert.Equal(t, test.expectedErr, err)
			assert.Equal(t, test.expectedResponse, gotResponse)
		})
	}
}

func TestFindUser(t *testing.T) {
	ctx := context.Background()
	addressBook := New()
	user := User

	findUserRequest := &pb.FindUserRequest{UserName: UserName}
	expectedResponse_without_error := &pb.FindUserResponse{Users: []*pb.User{user}}

	tests := map[string]struct {
		expectedResponse *pb.FindUserResponse
		expectedErr      error
	}{
		"without_error": {
			expectedResponse: expectedResponse_without_error,
			expectedErr:      nil,
		},
		"error": {
			expectedResponse: nil,
			expectedErr:      status.Errorf(codes.InvalidArgument, ErrNoSuchUser, findUserRequest.GetUserName()),
		},
	}

	for name, test := range tests {
		switch name {
		case "error":
			delete(addressBook.data, user.GetUserName())
		default:
			addressBook.data[user.GetUserName()] = user
		}
		t.Run(name, func(t *testing.T) {
			gotResponse, err := addressBook.FindUser(ctx, findUserRequest)
			assert.Equal(t, test.expectedErr, err)
			assert.Equal(t, test.expectedResponse, gotResponse)
		})

	}
}

func TestListUsers(t *testing.T) {
	ctx := context.Background()
	addressBook := New()
	user := User

	expectedErr := status.Error(codes.NotFound, ErrAddressBookIsEmpty)
	expectedResponse_without_error := &pb.ListUsersResponse{
		Users: []*pb.User{user},
	}

	tests := map[string]struct {
		expectedResponse *pb.ListUsersResponse
		expectedErr      error
	}{
		"without_error": {
			expectedResponse: expectedResponse_without_error,
			expectedErr:      nil,
		},
		"error": {
			expectedResponse: nil,
			expectedErr:      expectedErr,
		},
	}
	for name, test := range tests {
		switch name {
		case "error":
			delete(addressBook.data, user.GetUserName())
		default:
			addressBook.data[user.GetUserName()] = user
		}
		t.Run(name, func(t *testing.T) {
			gotListUsersResponse, err := addressBook.ListUsers(ctx, &empty.Empty{})
			assert.Equal(t, test.expectedErr, err)
			assert.Equal(t, test.expectedResponse, gotListUsersResponse)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()
	addressBook := New()
	user := User

	deleteUserRequest := &pb.DeleteUserRequest{UserName: UserName}
	expectedResponse_without_error := &pb.DeleteUserResponse{
		Response: fmt.Sprintf(DeleteUserMethodResponse, 1),
	}
	expectedErr := status.Errorf(codes.InvalidArgument, ErrNoSuchUser, deleteUserRequest.GetUserName())

	tests := map[string]struct {
		expectedResponse *pb.DeleteUserResponse
		expectedErr      error
	}{
		"without_error": {
			expectedResponse: expectedResponse_without_error,
			expectedErr:      nil,
		},
		"error": {
			expectedResponse: nil,
			expectedErr:      expectedErr,
		},
	}
	for name, test := range tests {
		switch name {
		case "error":
			delete(addressBook.data, user.GetUserName())
		default:
			addressBook.data[user.GetUserName()] = user
		}
		t.Run(name, func(t *testing.T) {
			gotResponse, err := addressBook.DeleteUser(ctx, deleteUserRequest)
			assert.Equal(t, test.expectedErr, err)
			assert.Equal(t, test.expectedResponse, gotResponse)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	ctx := context.Background()
	addressBook := New()
	user := User

	updatedUser := &pb.User{UserName: NewName, Phone: NewPhone, Address: NewAddress}
	updateUserRequest := &pb.UpdateUserRequest{
		UserName:    user.GetUserName(),
		UpdatedUser: updatedUser,
	}

	expectedResponse_without_error := &pb.UpdateUserResponse{
		Response: UpdateUserMethodResponse,
	}
	expectedErr := status.Errorf(codes.InvalidArgument, ErrNoSuchUser, updateUserRequest.GetUserName())

	tests := map[string]struct {
		expectedResponse *pb.UpdateUserResponse
		expectedErr      error
	}{
		"without_error": {
			expectedResponse: expectedResponse_without_error,
			expectedErr:      nil,
		},
		"error": {
			expectedResponse: nil,
			expectedErr:      expectedErr,
		},
	}

	for name, test := range tests {
		switch name {
		case "error":
			delete(addressBook.data, user.GetUserName())
		default:
			addressBook.data[user.GetUserName()] = user
		}
		t.Run(name, func(t *testing.T) {
			gotResponse, err := addressBook.UpdateUser(ctx, updateUserRequest)
			assert.Equal(t, test.expectedErr, err)
			assert.Equal(t, test.expectedResponse.GetResponse(), gotResponse.GetResponse())
		})
	}
}
