package main

import (
	"time"

	"infoblox-training-task-3/storage/pkg/dapr"
	"infoblox-training-task-3/storage/pkg/model"
	"infoblox-training-task-3/storage/pkg/pb"
	"infoblox-training-task-3/storage/pkg/svc"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/requestid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

func NewGRPCServer(logger *logrus.Logger, pubsub *dapr.PubSub, dbConnectionString string) (*grpc.Server, error) {
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

	// create new postgres database
	db, err := gorm.Open("postgres", dbConnectionString)
	if err != nil {
		return nil, err
	}
	if isInit := db.HasTable(&model.ResponderMode{}); !isInit {
		db.CreateTable(&model.ResponderMode{Mode: true})
	}

	// register service implementation with the grpcServer
	s, err := svc.NewBasicServer(pubsub, db, viper.GetString("app.id"), time.Now(), 0)
	if err != nil {
		return nil, err
	}
	pb.RegisterStorageServer(grpcServer, s)
	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	return grpcServer, nil
}
