package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/vstarostin/infoblox-training/infoblox-training-task-3/responder/pkg/dapr"
	"github.com/vstarostin/infoblox-training/infoblox-training-task-3/responder/pkg/model"
	"github.com/vstarostin/infoblox-training/infoblox-training-task-3/responder/pkg/pb"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	requestsCount       = 0
	modeFalseResponse   = "service is temporarily disabled"
	serviceUnavailable  = "service is unavailable"
	serviceRestarted    = "service restarted"
	errInvalidCommand   = "please, use commands: info, uptime, requests, mode, time or reset"
	errInvalidValue     = "please, use correct value"
	errTypeAssertion    = "type assertion error"
	errFailedToLoadData = "failed to load data"
)

type server struct {
	pubsub      *dapr.PubSub
	mu          sync.RWMutex
	description string
	startTime   time.Time
	requests    int64
	mode        bool
}

func (s *server) Handler(ctx context.Context, in *pb.HandlerRequest) (*pb.HandlerResponse, error) {
	s.mu.Lock()
	s.requests++
	s.mu.Unlock()

	var response string
	var err error
	if in.GetService() == pb.Service_RESPONDER {
		switch in.GetCommand() {
		case pb.Command_INFO:
			response = s.DescriptionHandler(in.Value)
		case pb.Command_UPTIME:
			in.Command = pb.Command_MODE
			response, err = s.GetUptime(in)
		case pb.Command_REQUESTS:
			response = s.GetRequestsCount()
		case pb.Command_MODE:
			response, err = s.ResponderModeStatus(in)
		case pb.Command_TIME:
			response = s.GetTime()
		case pb.Command_RESET:
			response, err = s.Reset(in)
		default:
			return nil, status.Error(codes.InvalidArgument, errInvalidCommand)
		}
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return &pb.HandlerResponse{Service: in.GetService().String(), Response: response}, nil
	}

	if in.GetService() == pb.Service_STORAGE && s.mode {
		id := uuid.New()

		message := &model.Message{
			ID:      id,
			Command: in.GetCommand().String(),
			Value:   in.GetValue(),
			Service: in.GetService().String(),
		}

		b, err := json.Marshal(message)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
		err = s.pubsub.Publish(viper.GetString("dapr.subscribe.topic"), b)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}

		select {
		case <-s.pubsub.Flag:
			value, ok := s.pubsub.IncomingData.LoadAndDelete(id)
			if !ok {
				return nil, status.Error(codes.Unknown, errFailedToLoadData)
			}
			response, ok := value.(string)
			if !ok {
				return nil, status.Error(codes.Unknown, errTypeAssertion)
			}
			return &pb.HandlerResponse{Service: in.GetService().String(), Response: response}, nil
		case <-time.After(5 * time.Second):
			return nil, status.Error(codes.Internal, serviceUnavailable)
		}
	}
	return nil, status.Error(codes.Internal, modeFalseResponse)
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	uptime := time.Since(s.startTime)
	return uptime.String(), nil
}

func (s *server) GetRequestsCount() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return strconv.Itoa(int(s.requests))
}

func (s *server) ResponderModeStatus(in *pb.HandlerRequest) (string, error) {
	if in.GetValue() != strconv.FormatBool(false) && in.GetValue() != strconv.FormatBool(true) && in.GetValue() != "" {
		return "", fmt.Errorf(errInvalidValue)
	}

	id := uuid.New()
	message := &model.Message{
		ID:      id,
		Command: in.GetCommand().String(),
		Value:   in.GetValue(),
		Service: in.GetService().String(),
	}

	b, err := json.Marshal(message)
	if err != nil {
		return "", err
	}
	err = s.pubsub.Publish(viper.GetString("dapr.subscribe.topic"), b)
	if err != nil {
		return "", err
	}

	select {
	case <-s.pubsub.Flag:
		value, ok := s.pubsub.IncomingData.LoadAndDelete(id)
		if !ok {
			return "", fmt.Errorf(errFailedToLoadData)
		}
		response, ok := value.(string)
		if !ok {
			return "", fmt.Errorf(errTypeAssertion)
		}

		mode, _ := strconv.ParseBool(response)
		s.mu.Lock()
		if mode != s.mode {
			s.mode = mode
			s.startTime = time.Now().UTC()
		}
		s.mu.Unlock()

		return response, nil
	case <-time.After(5 * time.Second):
		return "", fmt.Errorf(serviceUnavailable)
	}
}

func (s *server) GetTime() string {
	return time.Now().UTC().String()
}

func (s *server) Reset(in *pb.HandlerRequest) (string, error) {
	s.mu.Lock()
	s.description = viper.GetString("app.id")
	s.requests = requestsCount
	s.mu.Unlock()

	in.Command = pb.Command_MODE
	in.Value = strconv.FormatBool(true)
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

func NewBasicServer(pubsub *dapr.PubSub, description string, startTime time.Time, requests int64, mode bool) (pb.ResponderServer, error) {
	return &server{
		pubsub:      pubsub,
		description: description,
		startTime:   startTime,
		requests:    requests,
		mode:        mode,
	}, nil
}
