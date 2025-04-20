package app

import (
	"context"
	"net"
	"strings"

	"prx/internal/models"
	"prx/internal/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	pb.UnimplementedReverseServer
	app *App
}

// StartGRPC spins up the gRPC server on the given port.
func (a *App) startGRPC() {
	port := "50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		a.Log.Fatal("gRPC listen error", "err", err)
	}

	srv := grpc.NewServer(
		grpc.UnaryInterceptor((&grpcServer{app: a}).authInterceptor),
	)

	pb.RegisterReverseServer(srv, &grpcServer{app: a})
	a.Log.Info("gRPC server listening", "port", port)
	go func() {
		if err := srv.Serve(lis); err != nil {
			a.Log.Fatal("gRPC serve error", "err", err)
		}
	}()
}

func (s *grpcServer) authInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {

	md, _ := metadata.FromIncomingContext(ctx)
	auth := md["authorization"]
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

	if err := s.app.Kube.DeleteProxy(s.app.namespace, req.From); err != nil {
		return nil, err
	}
	s.app.deleteRedirectRecords(req.From)
	return &pb.Empty{}, nil
}

func (s *grpcServer) List(ctx context.Context, _ *pb.ListRequest) (*pb.ListResponse, error) {

	records, _ := s.app.getAllRedirectionRecords()
	resp := &pb.ListResponse{}
	for from, to := range records {
		resp.Records = append(resp.Records, &pb.ProxyRecord{From: from, To: to})
	}
	return resp, nil
}
