package main

import (
	"fmt"
	"time"

	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordsales/proto"
	"golang.org/x/net/context"
)

// GetStale gets the stale records
func (s *Server) GetStale(ctx context.Context, req *pb.GetStaleRequest) (*pb.GetStaleResponse, error) {
	config, err := s.load(ctx)
	if err != nil {
		return nil, err
	}
	resp := []*pb.Sale{}
	for _, sale := range config.Sales {
		if sale.Price <= 500 && time.Now().Sub(time.Unix(sale.LastUpdateTime, 0)) > time.Hour*24 {
			resp = append(resp, sale)
		}
	}

	return &pb.GetStaleResponse{StaleSales: resp}, err
}

// GetSaleState gets the state of a sale
func (s *Server) GetSaleState(ctx context.Context, req *pb.GetStateRequest) (*pb.GetStateResponse, error) {
	config, err := s.load(ctx)
	if err != nil {
		return nil, err
	}
	resp := []*pb.Sale{}
	for _, sale := range config.Sales {
		if sale.InstanceId == req.InstanceId {
			resp = append(resp, sale)
		}
	}

	for _, sale := range config.Archives {
		if sale.InstanceId == req.InstanceId {
			resp = append(resp, sale)
		}
	}

	return &pb.GetStateResponse{Sales: resp}, err
}

//ClientUpdate forces a move
func (s *Server) ClientUpdate(ctx context.Context, in *pbrc.ClientUpdateRequest) (*pbrc.ClientUpdateResponse, error) {
	return &pbrc.ClientUpdateResponse{}, s.syncSales(ctx, in.GetInstanceId())
}

func (s *Server) UpdatePrice(ctx context.Context, req *pb.UpdatePriceRequest) (*pb.UpdatePriceResponse, error) {
	config, err := s.load(ctx)
	if err != nil {
		return nil, err
	}

	price, err := s.getter.getPrice(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	if val, ok := config.PriceHistory[req.GetId()]; !ok {
		config.PriceHistory[req.GetId()] = &pb.Prices{History: []*pb.PriceHistory{&pb.PriceHistory{
			Date:  time.Now().Unix(),
			Price: price,
		}}}
	} else {
		latest := int64(0)
		value := float32(0)
		for _, h := range val.History {
			if h.Date > latest {
				latest = h.Date
				value = h.Price
			}
		}

		if price != value {
			val.History = append(val.History, &pb.PriceHistory{
				Date:  time.Now().Unix(),
				Price: price,
			})
		}
	}

	s.Log(fmt.Sprintf("%v", config.PriceHistory[req.GetId()]))

	return &pb.UpdatePriceResponse{}, s.save(ctx, config)
}
