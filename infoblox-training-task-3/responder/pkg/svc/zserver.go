package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"infoblox-training-task-3/responder/pkg/dapr"
	"infoblox-training-task-3/responder/pkg/pb"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// version is the current version of the service
	version = "0.0.1"
)

// Default implementation of the Responder server interface
type server struct {
	pubsub      *dapr.PubSub
	description string
	startTime   time.Time
	requests    int64
	mode        bool
}

// GetVersion returns the current version of the service
func (s *server) GetVersion(context.Context, *empty.Empty) (*pb.VersionResponse, error) {
	s.requests++
	return &pb.VersionResponse{Version: version}, nil
}

func (s *server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	s.requests++
	if in.GetService() == "responder" {
		switch in.GetCommand() {
		case "info":
			description := s.GetDescr(in.Value)
			return &pb.GetResponse{Service: in.GetService(), Response: description}, nil
		case "uptime":
			uptime := s.GetUpt()
			return &pb.GetResponse{Service: in.GetService(), Response: uptime}, nil
		case "requests":
			requests := s.GetRequests()
			return &pb.GetResponse{Service: in.GetService(), Response: requests}, nil
		case "mode":
			value, err := s.GetMode(in.Value)
			if err != nil {
				return nil, status.Error(codes.Unknown, "err")
			}
			return &pb.GetResponse{Service: in.GetService(), Response: strconv.FormatBool(value)}, nil
		case "time":
			time := s.GetTime()
			return &pb.GetResponse{Service: in.GetService(), Response: time}, nil
		case "reset":
			status := s.Reset()
			return &pb.GetResponse{Service: in.GetService(), Response: status}, nil
		default:
			return nil, status.Error(codes.InvalidArgument, "please provide commands: info, uptime, requests or reset")
		}
	}

	if in.GetService() == "storage" && s.mode {
		b, err := json.Marshal(in)
		if err != nil {
			return nil, status.Error(codes.Unknown, "err")
		}
		err = s.pubsub.Publish("topic3", b)
		if err != nil {
			return nil, status.Error(codes.Unknown, "error")
		}
		time.Sleep(10 * time.Millisecond)
		return &pb.GetResponse{Service: in.GetService(), Response: string(s.pubsub.IncomingData)}, nil
	}

	return nil, status.Error(codes.Unknown, "service is temporarily disabled")
}

func (s *server) GetDescr(value string) string {
	if value == "" {
		return s.description
	}
	s.description = value
	return s.description
}

func (s *server) GetUpt() string {
	if s.mode {
		uptime := time.Since(s.startTime)
		return uptime.String()
	}
	return "service is temporarily disabled"
}

func (s *server) GetRequests() string {
	return string(rune(s.requests))
}

func (s *server) GetMode(value string) (bool, error) {
	if value == "" {
		return s.mode, nil
	}
	if value == "false" || value == "true" {
		v, err := strconv.ParseBool(value)
		if err != nil {
			return false, err
		}
		s.mode = v
		s.startTime = time.Now()
		return s.mode, nil
	}
	return false, fmt.Errorf("err")
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
func NewBasicServer(pubsub *dapr.PubSub, description string, startTime time.Time, requests int64, mode bool) (pb.ResponderServer, error) {
	return &server{
		pubsub:      pubsub,
		description: description,
		startTime:   startTime,
		requests:    requests,
		mode:        mode,
	}, nil
}
