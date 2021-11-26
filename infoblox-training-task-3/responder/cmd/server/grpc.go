package main

import (
	"time"

	"github.com/vstarostin/infoblox-training/infoblox-training-task-3/responder/pkg/dapr"
	"github.com/vstarostin/infoblox-training/infoblox-training-task-3/responder/pkg/pb"
	"github.com/vstarostin/infoblox-training/infoblox-training-task-3/responder/pkg/svc"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/requestid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

func NewGRPCServer(logger *logrus.Logger, pubsub *dapr.PubSub) (*grpc.Server, error) {
	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(
			keepalive.ServerParameters{
				Time:    time.Duration(viper.GetInt("config.keepalive.time")) * time.Second,
				Timeout: time.Duration(viper.GetInt("config.keepalive.timeout")) * time.Second,
			},
		),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				// logging middleware
				grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logger)),

				// Request-Id interceptor
				requestid.UnaryServerInterceptor(),

				// Metrics middleware
				grpc_prometheus.UnaryServerInterceptor,

				// validation middleware
				grpc_validator.UnaryServerInterceptor(),

				// collection operators middleware
				gateway.UnaryServerInterceptor(),
			),
		),
	)

	// register service implementation with the grpcServer
	s, err := svc.NewBasicServer(pubsub, viper.GetString("app.id"), time.Now().UTC(), viper.GetInt64("app.requests"), viper.GetBool("app.mode"))
	if err != nil {
		return nil, err
	}
	pb.RegisterResponderServer(grpcServer, s)
	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	return grpcServer, nil
}
