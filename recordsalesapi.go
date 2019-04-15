package main

import (
	"time"

	pb "github.com/brotherlogic/recordsales/proto"
	"golang.org/x/net/context"
)

// GetStale gets the stale records
func (s *Server) GetStale(ctx context.Context, req *pb.GetStaleRequest) (*pb.GetStaleResponse, error) {
	resp := []*pb.Sale{}
	for _, sale := range s.config.Sales {
		if sale.Price <= 500 && time.Now().Sub(time.Unix(sale.LastUpdateTime, 0)) > time.Hour*24 {
			resp = append(resp, sale)
		}
	}

	return &pb.GetStaleResponse{StaleSales: resp}, nil
}

// GetSaleState gets the state of a sale
func (s *Server) GetSaleState(ctx context.Context, req *pb.GetStateRequest) (*pb.GetStateResponse, error) {
	resp := []*pb.Sale{}
	for _, sale := range s.config.Sales {
		if sale.InstanceId == req.InstanceId {
			resp = append(resp, sale)
		}
	}

	return &pb.GetStateResponse{Sales: resp}, nil
}
