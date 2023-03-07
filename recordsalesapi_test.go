package main

import (
	"log"
	"testing"
	"time"

	pbgd "github.com/brotherlogic/godiscogs/proto"
	pbrc "github.com/brotherlogic/recordcollection/proto"
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

func TestUpdatePrice(t *testing.T) {
	s := getTestServer()
	config := &pb.Config{}
	config.Sales = append(config.Sales, &pb.Sale{
		InstanceId:     12,
		Price:          123,
		LastUpdateTime: time.Now().Add(time.Hour * -48).Unix()})
	s.save(context.Background(), config)
	s.getter = &testGetter{records: []*pbrc.Record{&pbrc.Record{Metadata: &pbrc.ReleaseMetadata{
		SaleId:           12,
		CurrentSalePrice: 125,
		SaleState:        pbgd.SaleState_FOR_SALE,
		SaleBudget:       "madeup",
		Category:         pbrc.ReleaseMetadata_LISTED_TO_SELL}, Release: &pbgd.Release{InstanceId: 12}}}}

	_, err := s.ClientUpdate(context.Background(), &pbrc.ClientUpdateRequest{InstanceId: 12})
	if err != nil {
		log.Fatalf("Bad update: %v", err)
	}

	val, err := s.GetSaleState(context.Background(), &pb.GetStateRequest{InstanceId: 12})
	if err != nil {
		log.Fatalf("Bad get: %v", err)
	}

	if len(val.GetSales()) == 0 {
		t.Fatalf("Bad get sales: %v", val)
	}

	if val.GetSales()[0].GetPrice() == 123 {
		t.Errorf("Bad price: %v", val)
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
