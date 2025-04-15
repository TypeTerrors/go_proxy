package rpc

import (
	"context"
	"fmt"
	"prx/internal/models"
	"prx/internal/utils"
	pb "prx/proto"
	"time"
)

func (g *Grpc) AddNewProxy(ctx context.Context, req *pb.AddNewProxyRequest) (*pb.Response, error) {

	body := models.AddNewProxy{
		From: req.GetFrom(),
		To:   req.GetTo(),
		Cert: req.GetCert(),
		Key:  req.GetKey(),
	}

	if err := utils.ValidateFields(body); err != "" {
		return &pb.Response{
			Success: false,
			Error:   "both 'from' and 'to' fields are required",
		}, nil
	}

	g.Api.Log.Printf("Adding new proxy: from=%s to=%s", req.GetFrom(), req.GetTo())

	err := g.Api.Kube.AddNewProxy(body, g.Api.Namespace, g.Api.Name)
	if err != nil {
		return &pb.Response{Success: false, Error: err.Error()}, nil
	}

	g.Api.SetRedirectRecords(body.From, body.To)

	return &pb.Response{
		Success: true,
		Error:   "",
	}, nil
}

func (g *Grpc) DeleteProxy(ctx context.Context, req *pb.DeleteProxyRequest) (*pb.Response, error) {
	body := models.DelOldProxy{
		From: req.GetFrom(),
	}

	if errMsg := utils.ValidateFields(body); errMsg != "" {
		return &pb.Response{
			Success: false,
			Error:   "both 'from' and 'to' fields are required",
		}, nil
	}

	g.Api.Log.Printf("Deleting old proxy: from=%s", req.GetFrom())

	err := g.Api.Kube.DeleteProxy(g.Api.Namespace, body.From)
	if err != nil {
		return &pb.Response{Success: false, Error: err.Error()}, nil
	}

	g.Api.DeleteRedirectRecords(body.From)

	return &pb.Response{
		Success: true,
		Error:   "",
	}, nil
}

func (g *Grpc) PatchProxy(ctx context.Context, req *pb.PatchProxyRequest) (*pb.Response, error) {
	body := models.PatchOldProxy{
		From: req.GetFrom(),
		To:   req.GetTo(),
		Cert: req.GetCert(),
		Key:  req.GetKey(),
	}

	if errMsg := utils.ValidateFields(body); errMsg != "" {
		return &pb.Response{
			Success: false,
			Error:   "both 'from' and 'to' fields are required",
		}, nil
	}

	g.Api.Log.Printf("Patching old proxy: from=%s", req.GetFrom())

	err := g.Api.Kube.DeleteProxy(g.Api.Namespace, body.From)
	if err != nil {
		return &pb.Response{Success: false, Error: err.Error()}, nil
	}

	g.Api.DeleteRedirectRecords(body.From)

	err = g.Api.Kube.AddNewProxy(body, g.Api.Namespace, g.Api.Name)
	if err != nil {
		return &pb.Response{Success: false, Error: err.Error()}, nil
	}

	g.Api.SetRedirectRecords(body.From, body.To)

	return &pb.Response{
		Success: true,
		Error:   "",
	}, nil
}

func (g *Grpc) GetRedirectionRecords(ctx context.Context, req *pb.GetRedirectionRecordsRequest) (*pb.GetRedirectionRecordsResponse, error) {
	records, err := g.Api.GetAllRedirectionRecords()
	if err != nil {
		return &pb.GetRedirectionRecordsResponse{}, err
	}

	if len(records) < 1 {
		return &pb.GetRedirectionRecordsResponse{}, fmt.Errorf("no redirection records available")
	}

	var res []*pb.RedirectionRecord
	for i, v := range records {
		record := &pb.RedirectionRecord{
			From: i,
			To:   v,
		}
		res = append(res, record)
	}

	return &pb.GetRedirectionRecordsResponse{
		Records: res,
	}, nil
}

func (g *Grpc) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Status:  "OK",
		Time:    time.Now().Format(time.RFC3339),
		Version: g.Api.Version,
	}, nil
}
