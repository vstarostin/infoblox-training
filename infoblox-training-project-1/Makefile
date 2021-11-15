.PHONY: run
run:
	@export DB_CONNECTION_STRING=./secret/.db_conn && go run cmd/main.go

.PHONY: build
build:
	@go build -o ./app cmd/main.go

.PHONY: docker-build-generator
docker-build-generator:
	@docker build -t generator -f Dockerfile.generator .

.PHONE: gen-proto
gen-proto: docker-build-generator
	@docker run -d --rm \
		-v `pwd`/api:/api \
		-v `pwd`/internal/pb:/pb \
		generator protoc \
		--go_out=. --go_opt=paths=source_relative \
    	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. --grpc-gateway_opt=logtostderr=true --grpc-gateway_opt=paths=source_relative --grpc-gateway_opt=generate_unbound_methods=true \
		-I /usr/local/include/. \
    	-I /api/. api.proto

.PHONE: run-db
run-db: 
	@docker run -d --rm \
		--name db \
		-p 5432:5432 \
		-e POSTGRES_DB=postgres \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=password \
		postgres

.PHONY: docker-build
docker-build:
	@docker build -t address-book-service .

.PHONY: run-service
run-service:
	@docker run -d --rm \
		--name address-book-service \
		-p 8080:8080 \
		-v `pwd`/secret:/secret \
		-e DB_CONNECTION_STRING=./secret/.db_conn \
		address-book-service

.PHONY: docker-stop
docker-stop:
	@docker stop address-book-service

.PHONY: gen-mocks
gen-mocks:
	@docker run -v `pwd`:/src -w /src vektra/mockery:v2.7 --case snake --dir internal --output internal/mock --outpkg mock --all

.PHONY: compose-up	
compose-up:
	@docker-compose up -d 

.PHONY: compose-down	
compose-down:
	@docker-compose down

.PHONY: compose-start	
compose-start:
	@docker-compose start

.PHONY: compose-stop
compose-stop:
	@docker-compose stop


		