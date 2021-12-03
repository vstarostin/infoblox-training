package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/vstarostin/infoblox-training/infoblox-training-task-3/portal/pkg/pb"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/gorm/resource"
	"github.com/infobloxopen/atlas-app-toolkit/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func main() {
	doneC := make(chan error)
	logger := NewLogger()
	if viper.GetBool("internal.enable") {
		go func() { doneC <- ServeInternal(logger) }()
	}

	go func() { doneC <- ServeExternal(logger) }()

	if err := <-doneC; err != nil {
		logger.Fatal(err)
	}
}

func NewLogger() *logrus.Logger {
	logger := logrus.StandardLogger()
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logger.SetReportCaller(true)

	// Set the log level on the default logger based on command line flag
	if level, err := logrus.ParseLevel(viper.GetString("logging.level")); err != nil {
		logger.Errorf("Invalid %q provided for log level", viper.GetString("logging.level"))
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(level)
	}

	return logger
}

// ServeInternal builds and runs the server that listens on InternalAddress
func ServeInternal(logger *logrus.Logger) error {
	s, err := server.NewServer(
		// register metrics
		server.WithHandler("/metrics", promhttp.Handler()),
	)
	if err != nil {
		return err
	}
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", viper.GetString("internal.address"), viper.GetString("internal.port")))
	if err != nil {
		return err
	}

	logger.Debugf("serving internal http at %q", fmt.Sprintf("%s:%s", viper.GetString("internal.address"), viper.GetString("internal.port")))
	return s.Serve(nil, l)
}

// ServeExternal builds and runs the server that listens on ServerAddress and GatewayAddress
func ServeExternal(logger *logrus.Logger) error {
	grpcServer, err := NewGRPCServer(logger)
	if err != nil {
		logger.Fatalln(err)
	}
	grpc_prometheus.Register(grpcServer)

	s, err := server.NewServer(
		server.WithGrpcServer(grpcServer),
		server.WithGateway(
			gateway.WithGatewayOptions(
				runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
					MarshalOptions: protojson.MarshalOptions{
						UseProtoNames:   true,
						EmitUnpopulated: false,
					},
				}),
				runtime.WithForwardResponseOption(forwardResponseOption),
				runtime.WithIncomingHeaderMatcher(gateway.AtlasDefaultHeaderMatcher()),
			),
			gateway.WithServerAddress(fmt.Sprintf("%s:%s", viper.GetString("server.address"), viper.GetString("server.port"))),
			gateway.WithEndpointRegistration(viper.GetString("gateway.endpoint"), pb.RegisterPortalHandlerFromEndpoint),
		),
		server.WithHandler("/swagger/", NewSwaggerHandler(viper.GetString("gateway.swaggerFile"))),
	)
	if err != nil {
		logger.Fatalln(err)
	}

	grpcL, err := net.Listen("tcp", fmt.Sprintf("%s:%s", viper.GetString("server.address"), viper.GetString("server.port")))
	if err != nil {
		logger.Fatalln(err)
	}

	httpL, err := net.Listen("tcp", fmt.Sprintf("%s:%s", viper.GetString("gateway.address"), viper.GetString("gateway.port")))
	if err != nil {
		logger.Fatalln(err)
	}

	logger.Printf("serving gRPC at %s:%s", viper.GetString("server.address"), viper.GetString("server.port"))
	logger.Printf("serving http at %s:%s", viper.GetString("gateway.address"), viper.GetString("gateway.port"))

	return s.Serve(grpcL, httpL)
}

func init() {
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatalf("cannot bind flags: %v", err)
	}
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AddConfigPath(viper.GetString("config.source"))
	if viper.GetString("config.file") != "" {
		log.Printf("Serving from configuration file: %s", viper.GetString("config.file"))
		viper.SetConfigName(viper.GetString("config.file"))
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("cannot load configuration: %v", err)
		}
	} else {
		log.Printf("Serving from default values, environment variables, and/or flags")
	}
	resource.RegisterApplication(viper.GetString("app.id"))
	resource.SetPlural()
}

func forwardResponseOption(ctx context.Context, w http.ResponseWriter, resp protoreflect.ProtoMessage) error {
	w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
	return nil
}
