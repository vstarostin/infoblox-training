package handler

import (
	"context"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vstarostin/infoblox-training-project-1/internal/model"
	"github.com/vstarostin/infoblox-training-project-1/internal/pb"
)

type AddressBook struct {
	pb.UnimplementedAddressBookServiceServer
	service AddressBookService
}

type AddressBookService interface {
	AddUser(name, phone, address string) error
	ListUsers() ([]model.User, error)
	DeleteUser(name string) (string, error)
}

func New(service AddressBookService) *AddressBook {
	return &AddressBook{service: service}
}

func (ab *AddressBook) AddUser(_ context.Context, in *pb.AddUserRequest) (*pb.AddUserResponse, error) {
	name := format(in.GetNewUser().GetUserName())
	address := format(in.GetNewUser().GetAddress())
	phone := format(in.GetNewUser().GetPhone())

	err := ab.service.AddUser(name, phone, address)
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &pb.AddUserResponse{
		Response: "successfully added",
	}, nil
}

func (ab *AddressBook) ListUsers(_ context.Context, _ *empty.Empty) (*pb.ListUsersResponse, error) {
	users, err := ab.service.ListUsers()
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	response := &pb.ListUsersResponse{Users: make([]*pb.User, 0)}
	for _, u := range users {
		response.Users = append(response.Users, &pb.User{
			UserName: u.Name,
			Phone:    u.Phone,
			Address:  u.Address,
		})
	}

	return response, nil
}

func (ab *AddressBook) DeleteUser(_ context.Context, in *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	incomingNamePattern := format(in.GetUserName())
	response, err := ab.service.DeleteUser(incomingNamePattern)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.DeleteUserResponse{Response: response}, nil
}

func format(s string) string {
	cutset := " "
	return strings.ToLower(strings.Trim(s, cutset))
}
