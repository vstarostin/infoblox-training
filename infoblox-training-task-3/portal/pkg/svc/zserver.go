package svc

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/vstarostin/infoblox-training/infoblox-training-task-3/portal/pkg/pb"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	requestsCount      = 0
	serviceRestarted   = "service restarted"
	errInvalidArgument = "please use commands: info, uptime, requests or reset"
)

type server struct {
	Logger      *logrus.Logger
	grpcConn    *grpc.ClientConn
	mu          sync.RWMutex
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
	s.mu.Lock()
	s.requests++
	s.mu.Unlock()

	var response string
	if in.GetService() == pb.Service_PORTAL {
		switch in.GetCommand() {
		case pb.Command_INFO:
			response = s.DescriptionHandler(in.Value)
		case pb.Command_UPTIME:
			response = s.GetUptime()
		case pb.Command_REQUESTS:
			response = s.GetRequests()
		case pb.Command_TIME:
			response = s.GetTime()
		case pb.Command_RESET:
			response = s.Reset()
		default:
			return nil, status.Error(codes.InvalidArgument, errInvalidArgument)
		}
		return &pb.HandlerResponse{Service: in.GetService().String(), Response: response}, nil

	}
	if in.GetService() == pb.Service_STORAGE && in.GetCommand() == pb.Command_MODE {
		return nil, status.Error(codes.InvalidArgument, errInvalidArgument)
	}

	if in.GetService() == pb.Service_RESPONDER || in.GetService() == pb.Service_STORAGE {
		c := NewResponderClient(s.grpcConn)
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

func (s *server) GetDescription() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.description
}

func (s *server) SetDescription(value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.description = value
}

func (s *server) DescriptionHandler(value string) string {
	if value == "" {
		return s.GetDescription()
	}
	s.SetDescription(value)
	return s.GetDescription()
}

func (s *server) GetUptime() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	uptime := time.Since(s.startTime)
	return uptime.String()
}

func (s *server) GetRequests() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return strconv.Itoa(int(s.requests))
}

func (s *server) GetTime() string {
	return time.Now().UTC().String()
}

func (s *server) Reset() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.description = viper.GetString("app.id")
	s.requests = requestsCount
	s.startTime = time.Now().UTC()
	return serviceRestarted
}

func grpcConnection() (*grpc.ClientConn, error) {
	return grpc.Dial(fmt.Sprintf("%s:%s", viper.GetString("responder.address"), viper.GetString("responder.port")), grpc.WithInsecure())
}

// NewBasicServer returns an instance of the default server interface
func NewBasicServer(description string, startTime time.Time, requests int64) (pb.PortalServer, error) {
	conn, err := grpcConnection()
	if err != nil {
		return nil, err
	}

	return &server{
		description: description,
		startTime:   startTime,
		requests:    requests,
		grpcConn:    conn,
	}, nil
}
