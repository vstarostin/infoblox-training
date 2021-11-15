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

const (
	AddUserMethodResponse    = "successfully added"
	UpdateUserMethodResponse = "user was successfully updated"
	ErrUpdateUserMethod      = "please provide full phone number, address or name"
)

type AddressBook struct {
	pb.UnimplementedAddressBookServiceServer
	service AddressBookService
}

type AddressBookService interface {
	AddUser(name, phone, address string) error
	ListUsers() ([]model.User, error)
	DeleteUser(name string) (string, error)
	FindUser(name, phone, address string) ([]model.User, error)
	UpdateUser(phone string, updatedUser model.User) error
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
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}

	return &pb.AddUserResponse{
		Response: AddUserMethodResponse,
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

func (ab *AddressBook) FindUser(_ context.Context, in *pb.FindUserRequest) (*pb.FindUserResponse, error) {
	name := format(in.GetName())
	phone := format(in.GetPhone())
	address := format(in.GetAddress())

	usersFromDB, err := ab.service.FindUser(name, phone, address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var users []*pb.User
	for _, u := range usersFromDB {
		user := &pb.User{}
		user.UserName, user.Address, user.Phone = u.Name, u.Address, u.Phone
		users = append(users, user)
	}
	return &pb.FindUserResponse{Users: users}, nil
}

func (ab *AddressBook) UpdateUser(_ context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	phone := format(in.GetPhone())
	if strings.Contains(phone, "*") {
		return nil, status.Error(codes.InvalidArgument, ErrUpdateUserMethod)
	}
	newUserName := format(in.GetUpdatedUser().GetUserName())
	newAddress := format(in.GetUpdatedUser().GetAddress())
	newPhone := format(in.GetUpdatedUser().GetPhone())
	updatedUser := model.User{Name: newUserName, Phone: newPhone, Address: newAddress}
	err := ab.service.UpdateUser(phone, updatedUser)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.UpdateUserResponse{
		Response:    UpdateUserMethodResponse,
		UpdatedUser: in.GetUpdatedUser(),
	}, nil
}

func format(s string) string {
	return strings.ToLower(strings.Trim(s, " "))
}
