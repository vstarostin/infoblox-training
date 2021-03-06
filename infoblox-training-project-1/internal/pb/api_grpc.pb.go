// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pb

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// AddressBookServiceClient is the client API for AddressBookService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AddressBookServiceClient interface {
	AddUser(ctx context.Context, in *AddUserRequest, opts ...grpc.CallOption) (*AddUserResponse, error)
	FindUser(ctx context.Context, in *FindUserRequest, opts ...grpc.CallOption) (*FindUserResponse, error)
	DeleteUser(ctx context.Context, in *DeleteUserRequest, opts ...grpc.CallOption) (*DeleteUserResponse, error)
	ListUsers(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ListUsersResponse, error)
	UpdateUser(ctx context.Context, in *UpdateUserRequest, opts ...grpc.CallOption) (*UpdateUserResponse, error)
}

type addressBookServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAddressBookServiceClient(cc grpc.ClientConnInterface) AddressBookServiceClient {
	return &addressBookServiceClient{cc}
}

func (c *addressBookServiceClient) AddUser(ctx context.Context, in *AddUserRequest, opts ...grpc.CallOption) (*AddUserResponse, error) {
	out := new(AddUserResponse)
	err := c.cc.Invoke(ctx, "/pb.AddressBookService/AddUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *addressBookServiceClient) FindUser(ctx context.Context, in *FindUserRequest, opts ...grpc.CallOption) (*FindUserResponse, error) {
	out := new(FindUserResponse)
	err := c.cc.Invoke(ctx, "/pb.AddressBookService/FindUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *addressBookServiceClient) DeleteUser(ctx context.Context, in *DeleteUserRequest, opts ...grpc.CallOption) (*DeleteUserResponse, error) {
	out := new(DeleteUserResponse)
	err := c.cc.Invoke(ctx, "/pb.AddressBookService/DeleteUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *addressBookServiceClient) ListUsers(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ListUsersResponse, error) {
	out := new(ListUsersResponse)
	err := c.cc.Invoke(ctx, "/pb.AddressBookService/ListUsers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *addressBookServiceClient) UpdateUser(ctx context.Context, in *UpdateUserRequest, opts ...grpc.CallOption) (*UpdateUserResponse, error) {
	out := new(UpdateUserResponse)
	err := c.cc.Invoke(ctx, "/pb.AddressBookService/UpdateUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AddressBookServiceServer is the server API for AddressBookService service.
// All implementations must embed UnimplementedAddressBookServiceServer
// for forward compatibility
type AddressBookServiceServer interface {
	AddUser(context.Context, *AddUserRequest) (*AddUserResponse, error)
	FindUser(context.Context, *FindUserRequest) (*FindUserResponse, error)
	DeleteUser(context.Context, *DeleteUserRequest) (*DeleteUserResponse, error)
	ListUsers(context.Context, *empty.Empty) (*ListUsersResponse, error)
	UpdateUser(context.Context, *UpdateUserRequest) (*UpdateUserResponse, error)
	mustEmbedUnimplementedAddressBookServiceServer()
}

// UnimplementedAddressBookServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAddressBookServiceServer struct {
}

func (UnimplementedAddressBookServiceServer) AddUser(context.Context, *AddUserRequest) (*AddUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddUser not implemented")
}
func (UnimplementedAddressBookServiceServer) FindUser(context.Context, *FindUserRequest) (*FindUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FindUser not implemented")
}
func (UnimplementedAddressBookServiceServer) DeleteUser(context.Context, *DeleteUserRequest) (*DeleteUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUser not implemented")
}
func (UnimplementedAddressBookServiceServer) ListUsers(context.Context, *empty.Empty) (*ListUsersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUsers not implemented")
}
func (UnimplementedAddressBookServiceServer) UpdateUser(context.Context, *UpdateUserRequest) (*UpdateUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUser not implemented")
}
func (UnimplementedAddressBookServiceServer) mustEmbedUnimplementedAddressBookServiceServer() {}

// UnsafeAddressBookServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AddressBookServiceServer will
// result in compilation errors.
type UnsafeAddressBookServiceServer interface {
	mustEmbedUnimplementedAddressBookServiceServer()
}

func RegisterAddressBookServiceServer(s grpc.ServiceRegistrar, srv AddressBookServiceServer) {
	s.RegisterService(&AddressBookService_ServiceDesc, srv)
}

func _AddressBookService_AddUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AddressBookServiceServer).AddUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.AddressBookService/AddUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AddressBookServiceServer).AddUser(ctx, req.(*AddUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AddressBookService_FindUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FindUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AddressBookServiceServer).FindUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.AddressBookService/FindUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AddressBookServiceServer).FindUser(ctx, req.(*FindUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AddressBookService_DeleteUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AddressBookServiceServer).DeleteUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.AddressBookService/DeleteUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AddressBookServiceServer).DeleteUser(ctx, req.(*DeleteUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AddressBookService_ListUsers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AddressBookServiceServer).ListUsers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.AddressBookService/ListUsers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AddressBookServiceServer).ListUsers(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _AddressBookService_UpdateUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AddressBookServiceServer).UpdateUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.AddressBookService/UpdateUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AddressBookServiceServer).UpdateUser(ctx, req.(*UpdateUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AddressBookService_ServiceDesc is the grpc.ServiceDesc for AddressBookService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AddressBookService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.AddressBookService",
	HandlerType: (*AddressBookServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddUser",
			Handler:    _AddressBookService_AddUser_Handler,
		},
		{
			MethodName: "FindUser",
			Handler:    _AddressBookService_FindUser_Handler,
		},
		{
			MethodName: "DeleteUser",
			Handler:    _AddressBookService_DeleteUser_Handler,
		},
		{
			MethodName: "ListUsers",
			Handler:    _AddressBookService_ListUsers_Handler,
		},
		{
			MethodName: "UpdateUser",
			Handler:    _AddressBookService_UpdateUser_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api.proto",
}
