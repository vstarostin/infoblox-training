package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	"github.com/vstarostin/infoblox-training-project-1/internal/config"
	"github.com/vstarostin/infoblox-training-project-1/internal/pb"
	"github.com/vstarostin/infoblox-training-project-1/internal/service"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	addressBookService := service.New()
	grpcServer := grpc.NewServer()
	pb.RegisterAddressBookServiceServer(grpcServer, addressBookService)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		log.Printf("GRPC server is listening on :%d", cfg.GRPCPort)
		err := grpcServer.Serve(lis)
		if err != nil && err != grpc.ErrServerStopped {
			log.Fatal(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mux := runtime.NewServeMux()
	opt := []grpc.DialOption{grpc.WithInsecure()}
	err = pb.RegisterAddressBookServiceHandlerFromEndpoint(
		ctx, mux, fmt.Sprintf(":%d", cfg.GRPCPort), opt,
	)
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	go func() {
		log.Printf("gRPC gateway server is listenng on :%d\n", cfg.Port)
		err = server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(shutdown)

	<-shutdown
	log.Println("Shutdown signal received")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("GRPC gateway server shutdown error")
	}

	grpcServer.GracefulStop()
	log.Println("Server stopped gracefully")
}