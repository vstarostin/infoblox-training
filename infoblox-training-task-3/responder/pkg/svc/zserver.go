package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"infoblox-training-task-3/responder/pkg/dapr"
	"infoblox-training-task-3/responder/pkg/pb"

	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	modeFalseResponse  = "service is temporarily disabled"
	serviceUnavailable = "service is unavailable"
	serviceRestarted   = "service restarted"
	errInvalidCommand  = "please, use commands: info, uptime, requests, mode, time or reset"
	errInvalidValue    = "please, use correct value"
)

// Default implementation of the Responder server interface
type server struct {
	pubsub      *dapr.PubSub
	description string
	startTime   time.Time
	requests    int64
	mode        bool
}

func (s *server) Handler(ctx context.Context, in *pb.HandlerRequest) (*pb.HandlerResponse, error) {
	atomic.AddInt64(&s.requests, 1)
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
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return &pb.HandlerResponse{Service: in.GetService(), Response: response}, nil

	}

	if in.GetService() == "storage" && s.mode {
		b, err := json.Marshal(in)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
		err = s.pubsub.Publish(viper.GetString("dapr.subscribe.topic"), b)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
		incomingData := struct{ Response, Service string }{"", ""}

		select {
		case data := <-s.pubsub.IncomingData:
			err = json.Unmarshal(data, &incomingData)
			if err != nil {
				return nil, status.Error(codes.Unknown, err.Error())
			}
			return &pb.HandlerResponse{Service: in.GetService(), Response: incomingData.Response}, nil
		case <-time.After(5 * time.Second):
			s.mode = false
			return nil, status.Error(codes.Unknown, serviceUnavailable)
		}
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

func (s *server) GetUptime(in *pb.HandlerRequest) (string, error) {
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

func (s *server) ResponderModeStatus(in *pb.HandlerRequest) (string, error) {
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

	select {
	case data := <-s.pubsub.IncomingData:
		err = json.Unmarshal(data, &incomingData)
		if err != nil {
			return "", err
		}
		mode, err := strconv.ParseBool(incomingData.Response)
		if err != nil {
			return "", err
		}
		if s.mode != mode {
			s.startTime = time.Now().UTC()
		}
		s.mode = mode
		return incomingData.Response, nil
	case <-time.After(5 * time.Second):
		s.mode = false
		return "", fmt.Errorf(serviceUnavailable)
	}
}

func (s *server) GetTime() string {
	return time.Now().UTC().String()
}

func (s *server) Reset(in *pb.HandlerRequest) (string, error) {
	s.description = viper.GetString("app.id")
	s.requests = viper.GetInt64("app.requests")
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
