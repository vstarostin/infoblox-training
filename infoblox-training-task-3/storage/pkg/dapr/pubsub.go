package dapr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	daprpb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/grpc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type PubSub struct {
	client         daprpb.DaprClient
	Logger         *logrus.Logger
	TopicSubscribe string
	Name           string
	description    string
	startTime      time.Time
	incomingData   []byte
	requests       int64
}

func InitPubsub(topic string, pubsubName string, appPort int, grpcPort int, log *logrus.Logger) (*PubSub, error) {
	var err error
	init := false
	ps := &PubSub{
		Logger:         log,
		TopicSubscribe: topic,
		Name:           pubsubName,
		description:    viper.GetString("app.id"),
		incomingData:   make([]byte, 0),
		startTime:      time.Now(),
		requests:       0,
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
	p.requests++
	b, ok := e.Data.([]byte)
	if !ok {
		return false, err
	}
	in := struct {
		Command, Value, Service string
	}{
		"", "", "",
	}
	err = json.Unmarshal(b, &in)
	if err != nil {
		return false, err
	}
	switch in.Command {
	case "info":
		description := p.GetDescription(in.Value)
		p.Publish(viper.GetString("dapr.publish.topic"), []byte(description))
	case "uptime":
		uptime := p.GetUptime()
		p.Publish(viper.GetString("dapr.publish.topic"), []byte(uptime))
	case "requests":
		requests := p.GetRequestsCount()
		p.Publish(viper.GetString("dapr.publish.topic"), []byte(requests))
	case "time":
		time := p.GetTime()
		p.Publish(viper.GetString("dapr.publish.topic"), []byte(time))
	case "reset":
		status := p.Reset()
		p.Publish(viper.GetString("dapr.publish.topic"), []byte(status))
	}
	return false, err
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

func (p *PubSub) GetDescription(value string) string {
	if value == "" {
		return p.description
	}
	p.description = value
	return p.description
}

func (p *PubSub) GetUptime() string {
	uptime := time.Since(p.startTime)
	return uptime.String()
}

func (p *PubSub) GetRequestsCount() string {
	return string(rune(p.requests))
}

func (p *PubSub) GetTime() string {
	return time.Now().String()
}

func (p *PubSub) Reset() string {
	p.description = viper.GetString("app.id")
	p.requests = viper.GetInt64("app.requests")
	p.startTime = time.Now()
	return "service restarted"
}
