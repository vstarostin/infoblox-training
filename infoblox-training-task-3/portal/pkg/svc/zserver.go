package svc

import (
	"context"
	"fmt"
	"time"

	"infoblox-training-task-3/portal/pkg/pb"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	// version is the current version of the service
	version = "0.0.1"
)

// Default implementation of the Portal server interface
type server struct {
	Logger      *logrus.Logger
	description string
	startTime   time.Time
	requests    int64
}

type ResponderClient interface {
	GetVersion(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.VersionResponse, error)
	Get(ctx context.Context, in *pb.GetRequest, opts ...grpc.CallOption) (*pb.GetResponse, error)
}

type responderClient struct {
	cc grpc.ClientConnInterface
}

func NewResponderClient(cc grpc.ClientConnInterface) ResponderClient {
	return &responderClient{cc}
}

func (c *responderClient) GetVersion(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.VersionResponse, error) {
	out := new(pb.VersionResponse)
	err := c.cc.Invoke(ctx, "/responder.Responder/GetVersion", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *responderClient) Get(ctx context.Context, in *pb.GetRequest, opts ...grpc.CallOption) (*pb.GetResponse, error) {
	out := new(pb.GetResponse)
	err := c.cc.Invoke(ctx, "/responder.Responder/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetVersion returns the current version of the service
func (s *server) GetVersion(context.Context, *empty.Empty) (*pb.VersionResponse, error) {
	s.requests++
	return &pb.VersionResponse{Version: version}, nil
}

func (s *server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	s.requests++
	var response string
	if in.GetService() == "portal" {
		switch in.GetCommand() {
		case "info":
			response = s.GetDescription(in.Value)
		case "uptime":
			response = s.GetUptime()
		case "requests":
			response = s.GetRequests()
		case "time":
			response = s.GetTime()
		case "reset":
			response = s.Reset()
		default:
			return nil, status.Error(codes.InvalidArgument, "please provide commands: info, uptime, requests or reset")
		}
		return &pb.GetResponse{Service: in.GetService(), Response: response}, nil

	}
	if in.GetService() == "storage" && in.GetCommand() == "mode" {
		return nil, status.Error(codes.InvalidArgument, "please provide commands: info, uptime, requests or reset")
	}

	if in.GetService() == "responder" || in.GetService() == "storage" {
		// conn, err := grpc.Dial("0.0.0.1:9095", grpc.WithInsecure())
		conn, err := grpc.Dial(fmt.Sprintf("%s:%s", viper.GetString("server.address"), viper.GetString("responder.port")), grpc.WithInsecure())
		if err != nil {
			s.Logger.Fatalf("Failed to dial %s: %v", fmt.Sprintf("%s:%s", viper.GetString("server.address"), viper.GetString("responder.port")), err)
		}
		defer conn.Close()
		c := NewResponderClient(conn)
		res, err := c.Get(context.Background(), &pb.GetRequest{
			Value:   in.GetValue(),
			Command: in.GetCommand(),
			Service: in.GetService(),
		})
		if err != nil {
			return nil, status.Error(codes.Unknown, fmt.Sprintf("Failed to call RPC method GetDescription: %v", err))
			// s.Logger.Fatalf("Failed to call RPC method GetDescription: %v", err)
		}
		response := &pb.GetResponse{
			Service:  res.GetService(),
			Response: res.GetResponse(),
		}
		return response, nil
	}
	return nil, status.Error(codes.InvalidArgument, "err")
}

func (s *server) GetDescription(value string) string {
	if value == "" {
		return s.description
	}
	s.description = value
	return s.description
}

func (s *server) GetUptime() string {
	uptime := time.Since(s.startTime)
	return uptime.String()
}

func (s *server) GetRequests() string {
	return string(rune(s.requests))
}

func (s *server) GetTime() string {
	return time.Now().String()
}

func (s *server) Reset() string {
	s.description = viper.GetString("app.id")
	s.requests = 0
	s.startTime = time.Now()
	return "service restarted"
}

// NewBasicServer returns an instance of the default server interface
func NewBasicServer(description string, startTime time.Time, requests int64) (pb.PortalServer, error) {
	return &server{
		description: description,
		startTime:   startTime,
		requests:    requests,
	}, nil
}
