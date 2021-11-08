package service

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/vstarostin/infoblox-training-project-1/internal/pb"
)

const (
	ErrUserAlreadyExist      = "user %v already exists"
	ErrNoSuchUser            = "no such user with name: %v"
	ErrAddressBookIsEmpty    = "addressBook is empty"
	ErrNameIsTaken           = "name %v is already taken. Please choose different name"
	AddUserMethodResponse    = "successfully added"
	DeleteUserMethodResponse = "%d user(s) was(were) successfully deleted"
	UpdateUserMethodResponse = "successfully updated"
)

type AddressBook struct {
	pb.UnimplementedAddressBookServiceServer
	mu   sync.RWMutex
	data map[string]*pb.User
}

func New() *AddressBook {
	return &AddressBook{data: make(map[string]*pb.User)}
}

func (ab *AddressBook) AddUser(_ context.Context, in *pb.AddUserRequest) (*pb.AddUserResponse, error) {
	name := format(in.GetNewUser().GetUserName())
	_, ok := ab.isUserExist(name)
	if ok {
		return nil, status.Errorf(codes.AlreadyExists, ErrUserAlreadyExist, name)
	}

	address := format(in.GetNewUser().GetAddress())
	phone := format(in.GetNewUser().GetPhone())
	newUser := &pb.User{
		UserName: name,
		Phone:    phone,
		Address:  address,
	}

	ab.mu.Lock()
	ab.data[name] = newUser
	ab.mu.Unlock()

	return &pb.AddUserResponse{
		Response: AddUserMethodResponse,
	}, nil
}

func (ab *AddressBook) FindUser(_ context.Context, in *pb.FindUserRequest) (*pb.FindUserResponse, error) {
	var count int
	var users []*pb.User
	var listUserResponse *pb.ListUsersResponse
	var err error

	incomingNamePattern := format(in.GetUserName())

	if incomingNamePattern == "" || incomingNamePattern == "*" {
		listUserResponse, err = ab.ListUsers(context.Background(), &emptypb.Empty{})
		if err != nil {
			return nil, err
		}
		return &pb.FindUserResponse{
			Users: listUserResponse.GetUsers(),
		}, nil
	}

	ab.mu.RLock()
	for name, user := range ab.data {
		match, err := path.Match(incomingNamePattern, name)
		if err != nil {
			return nil, err
		}
		if !match {
			continue
		}
		users = append(users, user)
		count++
	}
	ab.mu.RUnlock()

	if count == 0 {
		return nil, status.Errorf(codes.InvalidArgument, ErrNoSuchUser, incomingNamePattern)
	}

	return &pb.FindUserResponse{
		Users: users,
	}, nil
}

func (ab *AddressBook) DeleteUser(_ context.Context, in *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	var count int
	incomingNamePattern := format(in.GetUserName())

	ab.mu.Lock()
	for name := range ab.data {
		match, err := path.Match(incomingNamePattern, name)
		if err != nil {
			return nil, err
		}
		if !match {
			continue
		}
		delete(ab.data, name)
		count++
	}
	ab.mu.Unlock()

	if count == 0 {
		return nil, status.Errorf(codes.InvalidArgument, ErrNoSuchUser, incomingNamePattern)
	}

	return &pb.DeleteUserResponse{
		Response: fmt.Sprintf(DeleteUserMethodResponse, count),
	}, nil
}

func (ab *AddressBook) ListUsers(_ context.Context, _ *empty.Empty) (*pb.ListUsersResponse, error) {
	if count := ab.count(); count == 0 {
		return nil, status.Error(codes.NotFound, ErrAddressBookIsEmpty)
	}

	var users []*pb.User
	ab.mu.RLock()
	for _, user := range ab.data {
		users = append(users, user)
	}
	ab.mu.RUnlock()

	return &pb.ListUsersResponse{
		Users: users,
	}, nil
}

func (ab *AddressBook) UpdateUser(_ context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	name := format(in.GetUserName())
	user, ok := ab.isUserExist(name)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, ErrNoSuchUser, name)
	}

	response := &pb.UpdateUserResponse{
		Response:    "",
		UpdatedUser: &pb.User{},
	}
	newUserName := format(in.GetUpdatedUser().GetUserName())
	newAddress := format(in.GetUpdatedUser().GetAddress())
	newPhone := format(in.GetUpdatedUser().GetPhone())

	if newUserName != "" {
		_, ok := ab.isUserExist(newUserName)
		if ok {
			return nil, status.Errorf(codes.InvalidArgument, ErrNameIsTaken, name)
		}
		response.UpdatedUser.UserName = newUserName
	} else {
		response.UpdatedUser.UserName = user.GetUserName()
	}

	if newAddress != "" {
		response.UpdatedUser.Address = newAddress
	} else {
		response.UpdatedUser.Address = user.GetAddress()
	}

	if newPhone != "" {
		response.UpdatedUser.Phone = newPhone
	} else {
		response.UpdatedUser.Phone = user.GetPhone()
	}

	ab.mu.Lock()
	delete(ab.data, name)
	ab.data[response.GetUpdatedUser().GetUserName()] = response.GetUpdatedUser()
	ab.mu.Unlock()
	response.Response = UpdateUserMethodResponse

	return response, nil
}

func (ab *AddressBook) isUserExist(name string) (*pb.User, bool) {
	ab.mu.RLock()
	user, ok := ab.data[name]
	ab.mu.RUnlock()

	if !ok {
		return nil, false
	}
	return user, true
}

func (ab *AddressBook) count() int {
	ab.mu.RLock()
	defer ab.mu.RUnlock()
	return len(ab.data)
}

func format(s string) string {
	cutset := " "
	return strings.ToLower(strings.Trim(s, cutset))
}
