package rpc

import (
	"fmt"
	"net"

	pb "prx/proto"

	"prx/internal/api"

	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
)

type Grpc struct {
	pb.UnimplementedProxyServiceServer
	grpcServer *grpc.Server
	listener   net.Listener
	Api        *api.Api
	port       string
}

func NewGrpc(api *api.Api) *Grpc {

	port := "50051"
	// Create TCP listener
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(fmt.Sprintf("failed to listen on port %s: %v", port, err))
	}

	return &Grpc{
		grpcServer: grpc.NewServer(),
		Api:        api,
		listener:   listener,
		port:       port,
	}
}

func (g *Grpc) Start(done chan<- error) {
	go func() {
		log.Infof("Starting gRPC server on port %s...", g.port)
		pb.RegisterProxyServiceServer(g.grpcServer, g)
		if err := g.grpcServer.Serve(g.listener); err != nil {
			done <- fmt.Errorf("failed to serve gRPCserver: %v", err)
		}
		done <- nil
	}()
}
