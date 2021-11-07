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

type AddressBook struct {
	pb.UnimplementedAddressBookServiceServer
	mu   sync.RWMutex
	data map[string]*pb.User
}

func New() *AddressBook {
	return &AddressBook{data: make(map[string]*pb.User)}
}

func (ab *AddressBook) AddUser(_ context.Context, in *pb.AddUserRequest) (*pb.AddUserResponse, error) {
	name := strings.ToLower(strings.Trim(in.GetNewUser().GetUserName(), " "))
	_, ok := ab.isUserExist(name)
	if ok {
		return nil, status.Errorf(codes.AlreadyExists, "user %v already exists", name)
	}

	ab.mu.Lock()
	ab.data[name] = in.GetNewUser()
	ab.mu.Unlock()

	return &pb.AddUserResponse{
		Response: "successfully added",
	}, nil
}

func (ab *AddressBook) FindUser(_ context.Context, in *pb.FindUserRequest) (*pb.FindUserResponse, error) {
	var count int
	var users []*pb.User
	var listUserResponse *pb.ListUsersResponse
	var err error

	incomingNamePattern := strings.ToLower(strings.Trim(in.GetUserName(), " "))

	if incomingNamePattern == "" || incomingNamePattern == "*" {
		listUserResponse, err = ab.ListUsers(context.Background(), &emptypb.Empty{})
		if err != nil {
			return nil, err
		}
		return &pb.FindUserResponse{
			Users: listUserResponse.GetUsers(),
		}, nil
	}

	if !strings.Contains(incomingNamePattern, "*") {
		user, ok := ab.isUserExist(incomingNamePattern)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "no such user with namepattern: %v", incomingNamePattern)
		}
		return &pb.FindUserResponse{
			Users: []*pb.User{user},
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
		return nil, status.Errorf(codes.InvalidArgument, "no such user with namepattern: %v", incomingNamePattern)
	}

	return &pb.FindUserResponse{
		Users: users,
	}, nil
}

func (ab *AddressBook) DeleteUser(_ context.Context, in *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	var count int
	incomingNamePattern := strings.ToLower(strings.Trim(in.GetUserName(), " "))

	if !strings.Contains(incomingNamePattern, "*") {
		_, ok := ab.isUserExist(incomingNamePattern)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "no such user with namepattern: %v", incomingNamePattern)
		}
		ab.deleteUser(incomingNamePattern)

		return &pb.DeleteUserResponse{
			Response: "user was successfully deleted",
		}, nil
	}

	ab.mu.RLock()
	for name := range ab.data {
		match, err := path.Match(incomingNamePattern, name)
		if err != nil {
			return nil, err
		}
		if !match {
			continue
		}
		ab.deleteUser(name)
		count++
	}
	ab.mu.RUnlock()

	if count == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "no such user with namepattern %v", incomingNamePattern)
	}

	return &pb.DeleteUserResponse{
		Response: fmt.Sprintf("%d user(s) was(were) successfully deleted", count),
	}, nil
}

func (ab *AddressBook) ListUsers(_ context.Context, _ *empty.Empty) (*pb.ListUsersResponse, error) {
	if count := ab.count(); count == 0 {
		return nil, status.Error(codes.NotFound, "addressBook is empty")
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

func (ab *AddressBook) deleteUser(key string) {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	delete(ab.data, key)
}