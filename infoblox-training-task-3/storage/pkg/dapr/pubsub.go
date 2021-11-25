package dapr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"infoblox-training-task-3/storage/pkg/pb"

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
	IncomingData   chan []byte
}

func InitPubsub(topic string, pubsubName string, appPort int, grpcPort int, log *logrus.Logger) (*PubSub, error) {
	var err error
	init := false
	ps := &PubSub{
		Logger:         log,
		TopicSubscribe: topic,
		Name:           pubsubName,
		IncomingData:   make(chan []byte),
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

	b, ok := e.Data.([]byte)
	if !ok {
		return false, err
	}
	in := struct {
		Command, Value, Service string
	}{
		"", "", "",
	}
	_ = json.Unmarshal(b, &in)
	if err != nil {
		return false, err
	}

	conn, err := grpc.Dial("127.0.0.1:9090", grpc.WithInsecure())
	if err != nil {
		p.Logger.Fatalf("Failed to dial %s: %v", "127.0.0.1:9090", err)
	}
	defer conn.Close()
	c := pb.NewStorageClient(conn)
	res, err := c.Get(context.Background(), &pb.GetRequest{
		Value:   in.Value,
		Command: in.Command,
		Service: in.Service,
	})
	if err != nil {
		return false, err
	}
	b, err = json.Marshal(res)
	if err != nil {
		return false, err
	}
	err = p.Publish(viper.GetString("dapr.publish.topic"), b)
	if err != nil {
		return false, err
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
