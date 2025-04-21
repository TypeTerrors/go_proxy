package app

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"net"
	"os"
	"strings"

	"prx/internal/models"
	"prx/internal/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	pb.UnimplementedReverseServer
	app *App
}

// StartGRPC spins up the gRPC server on the given port.
func (a *App) startGRPC() {
	port := ":50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		a.Log.Fatal("gRPC listen error", "err", err)
	}

	// TLS cert and key are provided as base64‑encoded env vars.
	crtB64 := os.Getenv("TLS_CRT")
	keyB64 := os.Getenv("TLS_KEY")
	if crtB64 == "" || keyB64 == "" {
		a.Log.Fatal("TLS_CRT or TLS_KEY environment variable not set")
	}

	crtPEM, err := base64.StdEncoding.DecodeString(crtB64)
	if err != nil {
		a.Log.Fatal("failed to decode TLS_CRT:", "err", err)
	}
	keyPEM, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		a.Log.Fatal("failed to decode TLS_KEY:", "err", err)
	}

	pair, err := tls.X509KeyPair(crtPEM, keyPEM)
	if err != nil {
		a.Log.Fatal("failed to load X509 key pair:", "err", err)
	}
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{pair},
	})
	
	srv := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor((&grpcServer{app: a}).authInterceptor),
	)

	pb.RegisterReverseServer(srv, &grpcServer{app: a})
	a.Log.Info("gRPC server listening", "port", port)
	if err := srv.Serve(lis); err != nil {
		a.Log.Fatal("gRPC serve error", "err", err)
	}
}

func (s *grpcServer) authInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {

	s.app.Log.Info("RPC incoming request", "req", req)

	md, _ := metadata.FromIncomingContext(ctx)
	auth := md["Authorization"]
	if len(auth) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization required")
	}
	token := strings.TrimPrefix(auth[0], "Bearer ")
	if _, err := s.app.Jwt.ValidateJWT(token); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}
	return handler(ctx, req)
}

func (s *grpcServer) Add(ctx context.Context, req *pb.ProxyRequest) (*pb.Empty, error) {

	s.app.Log.Info("RPC add new request", "req", req)

	err := s.app.Kube.AddNewProxy(models.AddNewProxy{
		From: req.From, To: req.To, Cert: req.Cert, Key: req.Key,
	}, s.app.namespace, s.app.name)
	if err != nil {
		return nil, err
	}
	s.app.setRedirectRecords(req.From, req.To)
	return &pb.Empty{}, nil
}

func (s *grpcServer) Update(ctx context.Context, req *pb.ProxyRequest) (*pb.Empty, error) {

	s.app.Log.Info("RPC update request", "req", req)

	if err := s.app.Kube.DeleteProxy(s.app.namespace, req.From); err != nil {
		return nil, err
	}
	s.app.deleteRedirectRecords(req.From)
	if err := s.app.Kube.AddNewProxy(models.AddNewProxy{
		From: req.From, To: req.To, Cert: req.Cert, Key: req.Key,
	}, s.app.namespace, s.app.name); err != nil {
		return nil, err
	}
	s.app.setRedirectRecords(req.From, req.To)
	return &pb.Empty{}, nil
}

func (s *grpcServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.Empty, error) {

	s.app.Log.Info("RPC delete request", "req", req)

	if err := s.app.Kube.DeleteProxy(s.app.namespace, req.From); err != nil {
		return nil, err
	}
	s.app.deleteRedirectRecords(req.From)
	return &pb.Empty{}, nil
}

func (s *grpcServer) List(ctx context.Context, _ *pb.ListRequest) (*pb.ListResponse, error) {

	s.app.Log.Info("RPC list request")

	records, _ := s.app.getAllRedirectionRecords()
	resp := &pb.ListResponse{}
	for from, to := range records {
		resp.Records = append(resp.Records, &pb.ProxyRecord{From: from, To: to})
	}
	return resp, nil
}
