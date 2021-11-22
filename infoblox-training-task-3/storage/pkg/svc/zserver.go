package svc

import (
	"context"
	"infoblox-training-task-3/storage/pkg/dapr"
	"infoblox-training-task-3/storage/pkg/pb"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// version is the current version of the service
	version = "0.0.1"
)

// Default implementation of the Storage server interface
type server struct {
	pubsub      *dapr.PubSub
	db          *gorm.DB
	description string
	startTime   time.Time
	requests    int64
}

// GetVersion returns the current version of the service
func (s *server) GetVersion(context.Context, *empty.Empty) (*pb.VersionResponse, error) {
	return &pb.VersionResponse{Version: version}, nil
}

func (s *server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	s.requests++
	var response string
	switch in.GetCommand() {
	case "info":
		response = s.GetDescription(in.GetValue())
	case "uptime":
		response = s.GetUptime()
	case "requests":
		response = s.GetRequestsCount()
	case "time":
		response = s.GetTime()
	case "reset":
		response = s.Reset()
	default:
		return nil, status.Error(codes.Unknown, "err")
	}
	if response == "" {
		return nil, status.Error(codes.Unknown, "err")
	}
	return &pb.GetResponse{
		Service:  in.GetService(),
		Response: response,
	}, nil
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

func (s *server) GetRequestsCount() string {
	return string(rune(s.requests))
}

func (s *server) GetTime() string {
	return time.Now().String()
}

func (s *server) Reset() string {
	s.description = viper.GetString("app.id")
	s.requests = viper.GetInt64("app.requests")
	s.startTime = time.Now()
	return "service restarted"
}

// NewBasicServer returns an instance of the default server interface
func NewBasicServer(pubsub *dapr.PubSub, database *gorm.DB, description string, startTime time.Time, requests int64) (pb.StorageServer, error) {
	return &server{
		pubsub:      pubsub,
		db:          database,
		description: description,
		startTime:   startTime,
		requests:    requests,
	}, nil
}
