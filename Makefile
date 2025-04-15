# Define the proto files
PROTO_FILES = ./proto/proxy.proto


# Define the Go plugin paths
PROTOC_GEN_GO=$(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC=$(shell go env GOPATH)/bin/protoc-gen-go-grpc

# Ensure the Go plugins are installed
install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Define the generate_proto target
generate_proto:
	@echo "Generating proto files..."
	protoc --proto_path=. --go_out=. --go-grpc_out=. $(PROTO_FILES)

# Define a clean target to remove generated files
clean:
	@echo "Cleaning generated files..."
	rm -rf ./proto/*.pb.go

# Run the Go application
# run: generate_proto
# 	go run ./cmd/server/main.go -sync-folder ./sync_folder -port 50051

# start seperate docker instances for testing
up:
	docker compose up -d
# stop and remove all docker instances of go_sync
down:
	docker compose down --rmi all

# restart all docker instances of go_sync
# restart:
# 	docker compose down --rmi all
# 	rm -rf sync_folder1 sync_folder2 sync_folder3 sync_folder4 sync_folder5 sync_folder6 sync_folder7 sync_folder8 sync_folder9
# 	docker compose up -d

# Default target
all: generate_proto

