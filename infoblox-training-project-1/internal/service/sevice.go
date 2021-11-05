package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	name := in.GetNewUser().GetUserName()
	if _, ok := ab.data[name]; ok {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("user %v already exists", name))
	}

	ab.mu.Lock()
	ab.data[name] = in.GetNewUser()
	ab.mu.Unlock()

	return &pb.AddUserResponse{
		Response: "successfully added",
	}, nil
}

func (ab *AddressBook) FindUser(_ context.Context, in *pb.FindUserRequest) (*pb.FindUserResponse, error) {
	name := in.GetUserName()

	ab.mu.RLock()
	user, ok := ab.data[name]
	ab.mu.RUnlock()
	if !ok {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	response := &pb.FindUserResponse{
		User: &pb.User{
			UserName: user.GetUserName(),
			Phone:    user.GetPhone(),
			Address:  user.GetAddress(),
		},
	}
	return response, nil
}

func (ab *AddressBook) DeleteUser(_ context.Context, in *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	name := in.GetUserName()
	if _, ok := ab.data[name]; !ok {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("user %v does not exist", name))
	}

	ab.mu.Lock()
	delete(ab.data, name)
	ab.mu.Unlock()

	return &pb.DeleteUserResponse{
		Response: "successfully deleted",
	}, nil
}

func (ab *AddressBook) ListUser(_ context.Context, _ *empty.Empty) (*pb.ListUserResponse, error) {
	if len(ab.data) == 0 {
		return nil, status.Error(codes.NotFound, "addressBook is empty")
	}

	var users []*pb.User
	ab.mu.RLock()
	for _, user := range ab.data {
		users = append(users, user)
	}
	ab.mu.RUnlock()

	return &pb.ListUserResponse{
		Users: users,
	}, nil
}
