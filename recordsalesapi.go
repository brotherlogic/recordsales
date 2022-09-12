package main

import (
	"fmt"
	"time"

	qpb "github.com/brotherlogic/queue/proto"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordsales/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
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
	rec, err := s.getter.loadRecord(ctx, in.GetInstanceId())
	if err != nil {
		if status.Convert(err).Code() == codes.OutOfRange {
			return &pbrc.ClientUpdateResponse{}, nil
		}
		return nil, err
	}

	err = s.syncSales(ctx, rec)
	if err != nil {
		return nil, err
	}

	conn, err := s.FDialServer(ctx, "queue")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	qclient := qpb.NewQueueServiceClient(conn)
	data, _ := proto.Marshal(&pb.UpdatePriceRequest{Id: rec.GetRelease().GetId()})
	_, err = qclient.AddQueueItem(ctx, &qpb.AddQueueItemRequest{
		QueueName:     "sale_update",
		RunTime:       time.Now().Add(time.Hour * 24 * 7).Unix(),
		Payload:       &google_protobuf.Any{Value: data},
		Key:           fmt.Sprintf("%v", rec.GetRelease().GetId()),
		RequireUnique: true,
	})

	return &pbrc.ClientUpdateResponse{}, err
}
func (s *Server) GetPrice(ctx context.Context, req *pb.GetPriceRequest) (*pb.GetPriceResponse, error) {
	config, err := s.load(ctx)
	if err != nil {
		return nil, err
	}
	prices := make(map[int32]*pb.Prices)

	for _, id := range req.GetIds() {
		if val, ok := config.PriceHistory[id]; ok {
			var latest *pb.PriceHistory
			for _, price := range val.GetHistory() {
				if latest == nil || price.GetDate() < latest.GetDate() {
					latest = price
				}
			}
			val.Latest = latest

			prices[id] = val
		} else {
			// Call out to rc to get the current price
			price, err := s.getter.getPrice(ctx, id)
			if err != nil {
				return nil, err
			}
			ph := &pb.PriceHistory{
				Date:  time.Now().Unix(),
				Price: price,
			}
			prices[id] = &pb.Prices{
				Latest:  ph,
				History: []*pb.PriceHistory{ph},
			}
		}
	}

	return &pb.GetPriceResponse{Prices: prices}, nil
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

	s.CtxLog(ctx, fmt.Sprintf("Got price for %v -> %v", req.GetId(), price))

	if val, ok := config.PriceHistory[req.GetId()]; !ok {
		config.PriceHistory[req.GetId()] = &pb.Prices{History: []*pb.PriceHistory{{
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

	s.CtxLog(ctx, fmt.Sprintf("%v", config.PriceHistory[req.GetId()]))

	err = s.save(ctx, config)
	if err != nil {
		return nil, err
	}

	conn, err := s.FDialServer(ctx, "queue")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	qclient := qpb.NewQueueServiceClient(conn)
	data, _ := proto.Marshal(req)
	_, err = qclient.AddQueueItem(ctx, &qpb.AddQueueItemRequest{
		QueueName: "sale_update",
		RunTime:   time.Now().Add(time.Hour * 24 * 7).Unix(),
		Payload:   &google_protobuf.Any{Value: data},
		Key:       fmt.Sprintf("%v", req.GetId()),
	})

	return &pb.UpdatePriceResponse{}, err
}
