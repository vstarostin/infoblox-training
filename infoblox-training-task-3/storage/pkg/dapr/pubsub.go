package dapr

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vstarostin/infoblox-training/infoblox-training-task-3/storage/pkg/svc"

	daprpb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/grpc"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

const requestsCount = 0

type PubSub struct {
	client         daprpb.DaprClient
	Logger         *logrus.Logger
	TopicSubscribe string
	Name           string
	svc            *svc.Service
}

func InitPubsub(topic string, pubsubName string, appPort int, grpcPort int, log *logrus.Logger, dbConnectionString string) (*PubSub, error) {
	init := false

	s, err := svc.NewService(viper.GetString("app.id"), dbConnectionString, time.Now().UTC(), requestsCount)
	if err != nil {
		return nil, err
	}

	ps := &PubSub{
		Logger:         log,
		TopicSubscribe: topic,
		Name:           pubsubName,
		svc:            s,
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

	b, err := p.svc.Handler(e.Data)
	if err != nil {
		return false, nil
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
