package svc

import (
	"context"
	"strconv"
	"time"

	"infoblox-training-task-3/storage/pkg/dapr"
	"infoblox-training-task-3/storage/pkg/model"
	"infoblox-training-task-3/storage/pkg/pb"

	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	serviceRestarted  = "service restarted"
	errInvalidCommand = "please, use commands: info, uptime, requests, mode, time or reset"
	errInvalidValue   = "please, use correct value"
)

// Default implementation of the Storage server interface
type server struct {
	pubsub      *dapr.PubSub
	db          *gorm.DB
	description string
	startTime   time.Time
	requests    int64
}

func (s *server) Handler(ctx context.Context, in *pb.HandlerRequest) (*pb.HandlerResponse, error) {
	s.requests++
	var response string
	var err error
	var mode bool
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
	case "mode":
		mode, err = s.ResponderModeStatus(in.GetValue())
		response = strconv.FormatBool(mode)
	default:
		return nil, status.Error(codes.InvalidArgument, errInvalidCommand)
	}
	if err != nil {
		return nil, status.Error(codes.Unknown, "err")
	}
	return &pb.HandlerResponse{
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
	return strconv.Itoa(int(s.requests))
}

func (s *server) GetTime() string {
	return time.Now().UTC().String()
}

func (s *server) GetMode(mode model.ResponderMode) bool {
	s.db.Find(&mode)
	return mode.Mode
}

func (s *server) SetMode(value bool) {
	s.db.Exec("UPDATE responder_modes SET mode=? WHERE id=?", value, 1)
}

func (s *server) ResponderModeStatus(in string) (bool, error) {
	responderMode := model.ResponderMode{}
	if in != "" {
		value, err := strconv.ParseBool(in)
		if err != nil {
			return false, err
		}
		s.SetMode(value)
	}
	return s.GetMode(responderMode), nil
}

func (s *server) Reset() string {
	s.description = viper.GetString("app.id")
	s.requests = viper.GetInt64("app.requests")
	s.startTime = time.Now().UTC()
	return serviceRestarted
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
