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
	// _ "github.com/lib/pq"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/vstarostin/infoblox-training-project-1/internal/config"
	"github.com/vstarostin/infoblox-training-project-1/internal/handler"
	"github.com/vstarostin/infoblox-training-project-1/internal/model"
	"github.com/vstarostin/infoblox-training-project-1/internal/pb"
	"github.com/vstarostin/infoblox-training-project-1/internal/repository"
	"github.com/vstarostin/infoblox-training-project-1/internal/service"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// dsn := "host=localhost user=postgres password=password dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(cfg.DBConnectionString), &gorm.Config{})
	if err != nil {
		log.Println("DB initializing error")
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	err = sqlDB.Ping()
	if err != nil {
		log.Println("DB pinging error")
		log.Fatal(err)
	}
	defer sqlDB.Close()
	log.Printf("Database connection successfully opened")

	// users := model.User{}
	db.AutoMigrate(&model.User{})
	log.Println("Database migrated")

	addressBookRepo := repository.New(db)
	addressBookService := service.New(addressBookRepo)
	addressBookHandler := handler.New(addressBookService)

	grpcServer := grpc.NewServer()
	pb.RegisterAddressBookServiceServer(grpcServer, addressBookHandler)

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
