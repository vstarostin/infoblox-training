package svc

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"infoblox-training-task-3/portal/pkg/pb"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	serviceRestarted   = "service restarted"
	errInvalidArgument = "please use commands: info, uptime, requests or reset"
)

// Default implementation of the Portal server interface
type server struct {
	Logger      *logrus.Logger
	description string
	startTime   time.Time
	requests    int64
}

type ResponderClient interface {
	Handler(ctx context.Context, in *pb.HandlerRequest, opts ...grpc.CallOption) (*pb.HandlerResponse, error)
}

type responderClient struct {
	cc grpc.ClientConnInterface
}

func NewResponderClient(cc grpc.ClientConnInterface) ResponderClient {
	return &responderClient{cc}
}

func (c *responderClient) Handler(ctx context.Context, in *pb.HandlerRequest, opts ...grpc.CallOption) (*pb.HandlerResponse, error) {
	out := &pb.HandlerResponse{}
	err := c.cc.Invoke(ctx, "/responder.Responder/Handler", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *server) Handler(ctx context.Context, in *pb.HandlerRequest) (*pb.HandlerResponse, error) {
	atomic.AddInt64(&s.requests, 1)
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
			return nil, status.Error(codes.InvalidArgument, errInvalidArgument)
		}
		return &pb.HandlerResponse{Service: in.GetService(), Response: response}, nil

	}
	if in.GetService() == "storage" && in.GetCommand() == "mode" {
		return nil, status.Error(codes.InvalidArgument, errInvalidArgument)
	}

	if in.GetService() == "responder" || in.GetService() == "storage" {
		conn, err := grpc.Dial(fmt.Sprintf("%s:%s", viper.GetString("server.address"), viper.GetString("responder.port")), grpc.WithInsecure())
		if err != nil {
			s.Logger.Fatalf("Failed to dial %s: %v", fmt.Sprintf("%s:%s", viper.GetString("server.address"), viper.GetString("responder.port")), err)
		}
		defer conn.Close()
		c := NewResponderClient(conn)
		res, err := c.Handler(context.Background(), &pb.HandlerRequest{
			Value:   in.GetValue(),
			Command: in.GetCommand(),
			Service: in.GetService(),
		})
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
		response := &pb.HandlerResponse{
			Service:  res.GetService(),
			Response: res.GetResponse(),
		}
		return response, nil
	}
	return nil, status.Error(codes.InvalidArgument, errInvalidArgument)
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
	return strconv.Itoa(int(s.requests))
}

func (s *server) GetTime() string {
	return time.Now().UTC().String()
}

func (s *server) Reset() string {
	s.description = viper.GetString("app.id")
	s.requests = 0
	s.startTime = time.Now().UTC()
	return serviceRestarted
}

// NewBasicServer returns an instance of the default server interface
func NewBasicServer(description string, startTime time.Time, requests int64) (pb.PortalServer, error) {
	return &server{
		description: description,
		startTime:   startTime,
		requests:    requests,
	}, nil
}
