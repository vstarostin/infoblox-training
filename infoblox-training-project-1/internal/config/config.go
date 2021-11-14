package config

import (
	"flag"
	"fmt"
)

type Config struct {
	Port               int
	GRPCPort           int
	DBConnectionString string
}

func NewConfig() *Config {
	port := flag.Int("port", 8080, "GRPC gateway server port")
	gRPCPort := flag.Int("grpcport", 9090, "GRPC server port")
	username := flag.String("username", "postgres", "database user")
	password := flag.String("password", "password", "database password")
	host := flag.String("host", "localhost", "database host")
	dbPort := flag.Int("dbport", 5432, "database port")
	dbName := flag.String("dbname", "postgres", "database name")
	flag.Parse()

	dbConn := fmt.Sprintf("postgresql://%v:%v@%v:%v/%v", *username, *password, *host, *dbPort, *dbName)

	return &Config{
		Port:               *port,
		GRPCPort:           *gRPCPort,
		DBConnectionString: dbConn,
	}
}
