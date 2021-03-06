package main

import (
	"testing"
	"time"

	pb "github.com/brotherlogic/recordsales/proto"
	"golang.org/x/net/context"
)

func TestListStale(t *testing.T) {
	s := getTestServer()
	config := &pb.Config{}
	config.Sales = append(config.Sales, &pb.Sale{InstanceId: 12, LastUpdateTime: time.Now().Add(time.Hour * -48).Unix()})
	config.Sales = append(config.Sales, &pb.Sale{InstanceId: 13, LastUpdateTime: time.Now().Add(time.Hour * -5).Unix()})
	s.save(context.Background(), config)

	resp, err := s.GetStale(context.Background(), &pb.GetStaleRequest{})
	if err != nil {
		t.Errorf("Problem with getting stale: %v", err)
	}
	if len(resp.StaleSales) != 1 || resp.StaleSales[0].InstanceId != 12 {
		t.Errorf("Bad stales: %v", resp)
	}
}

func TestGetState(t *testing.T) {
	s := getTestServer()
	config := &pb.Config{}
	config.Sales = append(config.Sales, &pb.Sale{InstanceId: 12, LastUpdateTime: time.Now().Add(time.Hour * -48).Unix()})
	config.Sales = append(config.Sales, &pb.Sale{InstanceId: 13, LastUpdateTime: time.Now().Add(time.Hour * -5).Unix()})
	s.save(context.Background(), config)

	resp, err := s.GetSaleState(context.Background(), &pb.GetStateRequest{InstanceId: 12})
	if err != nil {
		t.Errorf("Problem with getting stale: %v", err)
	}

	if len(resp.GetSales()) != 1 {
		t.Errorf("Bad get sales pull: %v", resp)
	}
}

func TestGetStateFromArchives(t *testing.T) {
	s := getTestServer()
	config := &pb.Config{}
	config.Archives = append(config.Sales, &pb.Sale{InstanceId: 12, LastUpdateTime: time.Now().Add(time.Hour * -48).Unix()})
	s.save(context.Background(), config)

	resp, err := s.GetSaleState(context.Background(), &pb.GetStateRequest{InstanceId: 12})
	if err != nil {
		t.Errorf("Problem with getting stale: %v", err)
	}

	if len(resp.GetSales()) == 1 {
		t.Errorf("Bad get sales pull: %v", resp)
	}
}
