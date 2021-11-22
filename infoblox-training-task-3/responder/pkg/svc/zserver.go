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
	version           = "0.0.1"
	modeFalseResponse = "service is temporarily disabled"
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
	var response string
	if in.GetService() == "responder" {
		switch in.GetCommand() {
		case "info":
			response = s.GetDescription(in.Value)
		case "uptime":
			response = s.GetUptime()
		case "requests":
			response = s.GetRequestsCount()
		case "mode":
			value, err := s.GetMode(in.Value)
			if err != nil {
				return nil, status.Error(codes.Unknown, "err")
			}
			response = strconv.FormatBool(value)
		case "time":
			response = s.GetTime()
		case "reset":
			response = s.Reset()
		default:
			return nil, status.Error(codes.InvalidArgument, "please provide commands: info, uptime, requests or reset")
		}
		return &pb.GetResponse{Service: in.GetService(), Response: response}, nil

	}

	if in.GetService() == "storage" && s.mode {
		b, err := json.Marshal(in)
		if err != nil {
			return nil, status.Error(codes.Unknown, "err")
		}
		err = s.pubsub.Publish(viper.GetString("dapr.subscribe.topic"), b)
		if err != nil {
			return nil, status.Error(codes.Unknown, "error")
		}
		time.Sleep(10 * time.Millisecond)
		incomingData := struct{ Response, Service string }{"", ""}
		err = json.Unmarshal(s.pubsub.IncomingData, &incomingData)
		if err != nil {
			return nil, status.Error(codes.Unknown, "err")
		}
		return &pb.GetResponse{Service: in.GetService(), Response: incomingData.Response}, nil
	}

	return nil, status.Error(codes.Unknown, modeFalseResponse)
}

func (s *server) GetDescription(value string) string {
	if value == "" {
		return s.description
	}
	s.description = value
	return s.description
}

func (s *server) GetUptime() string {
	if s.mode {
		uptime := time.Since(s.startTime)
		return uptime.String()
	}
	return modeFalseResponse
}

func (s *server) GetRequestsCount() string {
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
