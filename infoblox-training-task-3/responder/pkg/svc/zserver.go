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
	serviceRestarted  = "service restarted"
	errInvalidCommand = "please, use commands: info, uptime, requests, mode, time or reset"
	errInvalidValue   = "please, use correct value"
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
	var err error
	if in.GetService() == "responder" {
		switch in.GetCommand() {
		case "info":
			response = s.GetDescription(in.Value)
		case "uptime":
			in.Command = "mode"
			response, err = s.GetUptime(in)
		case "requests":
			response = s.GetRequestsCount()
		case "mode":
			response, err = s.ResponderModeStatus(in)
		case "time":
			response = s.GetTime()
		case "reset":
			response, err = s.Reset(in)
		default:
			return nil, status.Error(codes.InvalidArgument, errInvalidCommand)
		}
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, errInvalidValue)
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
		incomingData := struct{ Response, Service string }{"", ""}

		err = json.Unmarshal(<-s.pubsub.IncomingData, &incomingData)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
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

func (s *server) GetUptime(in *pb.GetRequest) (string, error) {
	if in.GetValue() != "" {
		return "", fmt.Errorf(errInvalidValue)
	}
	val, err := s.ResponderModeStatus(in)
	if err != nil {
		return "", err
	}
	value, err := strconv.ParseBool(val)
	if err != nil {
		return "", err
	}
	if !value {
		return modeFalseResponse, nil
	}
	uptime := time.Since(s.startTime)
	return uptime.String(), nil
}

func (s *server) GetRequestsCount() string {
	return strconv.Itoa(int(s.requests))
}

func (s *server) ResponderModeStatus(in *pb.GetRequest) (string, error) {
	if in.GetValue() != "false" && in.GetValue() != "true" && in.GetValue() != "" {
		return "", fmt.Errorf(errInvalidValue)
	}
	b, err := json.Marshal(in)
	if err != nil {
		return "", err
	}
	err = s.pubsub.Publish(viper.GetString("dapr.subscribe.topic"), b)
	if err != nil {
		return "", err
	}
	incomingData := struct{ Response, Service string }{"", ""}
	err = json.Unmarshal(<-s.pubsub.IncomingData, &incomingData)
	if err != nil {
		return "", err
	}
	mode, err := strconv.ParseBool(incomingData.Response)
	if err != nil {
		return "", err
	}
	if s.mode != mode {
		s.startTime = time.Now()
	}
	s.mode = mode
	return incomingData.Response, nil
}

func (s *server) GetTime() string {
	return time.Now().String()
}

func (s *server) Reset(in *pb.GetRequest) (string, error) {
	s.description = viper.GetString("app.id")
	s.requests = 0
	in.Command = "mode"
	in.Value = "true"
	val, err := s.ResponderModeStatus(in)
	if err != nil {
		return "", err
	}
	_, err = strconv.ParseBool(val)
	if err != nil {
		return "", err
	}
	return serviceRestarted, nil
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
