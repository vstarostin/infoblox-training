module github.com/vstarostin/infoblox-training/infoblox-training-task-3/storage

go 1.16

replace github.com/spf13/afero => github.com/spf13/afero v1.5.1

require (
	github.com/dapr/dapr v1.4.3
	github.com/dapr/go-sdk v1.3.0
	github.com/denisenkom/go-mssqldb v0.9.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0 // indirect
	github.com/infobloxopen/atlas-app-toolkit v1.1.1
	github.com/jinzhu/gorm v1.9.16
	github.com/jinzhu/now v1.1.1 // indirect
	github.com/mattn/go-sqlite3 v1.14.6 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.9.0
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.27.1
)
