package dapr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/vstarostin/infoblox-training/infoblox-training-task-3/storage/pkg/model"

	daprpb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/grpc"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

const (
	requestsCount                                = 0
	info, uptime, requests, timeStr, reset, mode = "INFO", "UPTIME", "REQUESTS", "TIME", "RESET", "MODE"
	errTypeAssertion                             = "type assertion error"
	errInvalidCommand                            = "please, use commands info, uptime, requests, mode, time or reset"
	serviceRestarted                             = "service restarted"
)

type PubSub struct {
	mu             sync.RWMutex
	client         daprpb.DaprClient
	db             *gorm.DB
	Logger         *logrus.Logger
	TopicSubscribe string
	Name           string
	description    string
	startTime      time.Time
	requests       int64
}

func InitPubsub(topic string, pubsubName string, appPort int, grpcPort int, log *logrus.Logger, dbConnectionString string) (*PubSub, error) {
	init := false
	db, err := gorm.Open("postgres", dbConnectionString)
	if err != nil {
		return nil, err
	}
	if isInit := db.HasTable(&model.ResponderMode{}); !isInit {
		db.CreateTable(&model.ResponderMode{Mode: true})
	}
	ps := &PubSub{
		Logger:         log,
		db:             db,
		TopicSubscribe: topic,
		Name:           pubsubName,
		description:    viper.GetString("app.id"),
		startTime:      time.Now().UTC(),
		requests:       requestsCount,
	}

	if pubsubName != "" && topic != "" && grpcPort >= 1 {
		ps.initSubscriber(appPort)
		init = true
	}

	if appPort >= 1 {
		if ps.client, err = ps.initPublisher(grpcPort); err != nil {
			return nil, err
		}
		init = true
	}

	if init {
		return ps, nil
	}
	return nil, fmt.Errorf("pubsub disabled")
}

func (p *PubSub) initPublisher(port int) (daprpb.DaprClient, error) {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to open atlas pubsub connection: %v", err)
	}
	return daprpb.NewDaprClient(conn), nil
}

func (p *PubSub) initSubscriber(appPort int) {
	s, err := daprd.NewService(fmt.Sprintf(":%d", appPort))
	if err != nil {
		p.Logger.Fatalf("failed to start the server: %v", err)
	}

	subscription := &common.Subscription{
		PubsubName: p.Name,
		Topic:      p.TopicSubscribe,
	}
	if err := s.AddTopicEventHandler(subscription, p.eventHandler); err != nil {
		p.Logger.Fatalf("error adding handler: %v", err)
	}

	// start the server to handle incoming events
	go func(service common.Service) {
		if err := service.Start(); err != nil {
			p.Logger.Fatalf("server error: %v", err)
		}
	}(s)
}

func (p *PubSub) eventHandler(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	p.Logger.Debugf("Incoming message from pubsub %q, topic %q, data: %s", e.PubsubName, e.Topic, e.Data)

	p.mu.Lock()
	p.requests++
	p.mu.Unlock()

	b, ok := e.Data.([]byte)
	if !ok {
		return false, fmt.Errorf(errTypeAssertion)
	}

	in := &model.Message{}
	err = json.Unmarshal(b, &in)
	if err != nil {
		return false, err
	}

	var response string
	switch in.Command {
	case info:
		response = p.HandlerDescription(in.Value)
	case uptime:
		response = p.GetUptime()
	case requests:
		response = p.GetRequestsCount()
	case timeStr:
		response = p.GetTime()
	case reset:
		response = p.Reset()
	case mode:
		mode := p.ResponderModeStatus(in.Value)
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
		return true, err
	}
	err = p.Publish(viper.GetString("dapr.publish.topic"), b)
	if err != nil {
		return true, err
	}
	return false, nil
}

func (p *PubSub) Publish(topic string, msg []byte) error {
	if p.client == nil {
		return errors.New("client is not initialized")
	}

	_, err := p.client.PublishEvent(context.Background(), &daprpb.PublishEventRequest{
		Topic:      topic,
		Data:       msg,
		PubsubName: p.Name,
	})
	return err
}

func (p *PubSub) SetDescription(value string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.description = value
}

func (p *PubSub) GetDescription() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.description
}

func (p *PubSub) HandlerDescription(value string) string {
	if value == "" {
		return p.GetDescription()
	}
	p.SetDescription(value)
	return p.GetDescription()
}

func (p *PubSub) GetUptime() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	uptime := time.Since(p.startTime)
	return uptime.String()
}

func (p *PubSub) GetRequestsCount() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return strconv.Itoa(int(p.requests))
}

func (p *PubSub) GetTime() string {
	return time.Now().UTC().String()
}

func (p *PubSub) GetMode(mode model.ResponderMode) bool {
	p.db.Find(&mode)
	return mode.Mode
}

func (p *PubSub) SetMode(value bool) {
	p.db.Exec("UPDATE responder_modes SET mode=? WHERE id=?", value, 1)
}

func (p *PubSub) ResponderModeStatus(in string) bool {
	responderMode := model.ResponderMode{}
	if in != "" {
		value, err := strconv.ParseBool(in)
		if err != nil {
			return false
		}
		p.SetMode(value)
	}
	return p.GetMode(responderMode)
}

func (p *PubSub) Reset() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.description = viper.GetString("app.id")
	p.requests = requestsCount
	p.startTime = time.Now().UTC()
	return serviceRestarted
}
