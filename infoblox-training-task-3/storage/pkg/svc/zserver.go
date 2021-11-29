package svc

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"github.com/vstarostin/infoblox-training/infoblox-training-task-3/storage/pkg/model"
)

const (
	requestsCount          = 0
	errTypeAssertion       = "type assertion error"
	info, uptime, requests = "INFO", "UPTIME", "REQUESTS"
	timeStr, reset, mode   = "TIME", "RESET", "MODE"
	errInvalidCommand      = "please, use commands info, uptime, requests, mode, time or reset"
	serviceRestarted       = "service restarted"
)

type Service struct {
	mu          sync.RWMutex
	db          *gorm.DB
	description string
	startTime   time.Time
	requests    int64
}

func NewService(description, dbConnectionString string, startTime time.Time, requests int64) (*Service, error) {
	db, err := gorm.Open("postgres", dbConnectionString)
	if err != nil {
		return nil, err
	}
	if isInit := db.HasTable(&model.ResponderMode{}); !isInit {
		db.CreateTable(&model.ResponderMode{Mode: true})
	}
	return &Service{
		db:          db,
		description: description,
		startTime:   time.Now().UTC(),
		requests:    requestsCount,
	}, nil
}

func (s *Service) Handler(e interface{}) ([]byte, error) {
	s.mu.Lock()
	s.requests++
	s.mu.Unlock()

	b, ok := e.([]byte)
	if !ok {
		return []byte{}, fmt.Errorf(errTypeAssertion)
	}

	in := &model.Message{}
	err := json.Unmarshal(b, &in)
	if err != nil {
		return []byte{}, err
	}

	var response string
	switch in.Command {
	case info:
		response = s.HandlerDescription(in.Value)
	case uptime:
		response = s.GetUptime()
	case requests:
		response = s.GetRequestsCount()
	case timeStr:
		response = s.GetTime()
	case reset:
		response = s.Reset()
	case mode:
		mode := s.ResponderModeStatus(in.Value)
		response = strconv.FormatBool(mode)
	default:
		response = errInvalidCommand
	}

	resp := &model.MessagePubSub{
		ID:       in.ID,
		Service:  in.Service,
		Response: response,
	}
	b, err = json.Marshal(resp)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}

func (s *Service) SetDescription(value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.description = value
}

func (s *Service) GetDescription() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.description
}

func (s *Service) HandlerDescription(value string) string {
	if value == "" {
		return s.GetDescription()
	}
	s.SetDescription(value)
	return s.GetDescription()
}

func (s *Service) GetUptime() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	uptime := time.Since(s.startTime)
	return uptime.String()
}

func (s *Service) GetRequestsCount() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return strconv.Itoa(int(s.requests))
}

func (s *Service) GetTime() string {
	return time.Now().UTC().String()
}

func (s *Service) GetMode(mode model.ResponderMode) bool {
	s.db.Find(&mode)
	return mode.Mode
}

func (s *Service) SetMode(value bool) {
	s.db.Exec("UPDATE responder_modes SET mode=? WHERE id=?", value, 1)
}

func (s *Service) ResponderModeStatus(in string) bool {
	responderMode := model.ResponderMode{}
	if in != "" {
		value, err := strconv.ParseBool(in)
		if err != nil {
			return false
		}
		s.SetMode(value)
	}
	return s.GetMode(responderMode)
}

func (s *Service) Reset() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.description = viper.GetString("app.id")
	s.requests = requestsCount
	s.startTime = time.Now().UTC()
	return serviceRestarted
}
