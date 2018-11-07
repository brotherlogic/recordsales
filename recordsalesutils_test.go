package main

import (
	"fmt"
	"testing"

	"github.com/brotherlogic/keystore/client"
	"golang.org/x/net/context"

	pbgd "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordsales/proto"
)

type testGetter struct {
	records []*pbrc.Record
	fail    bool
}

func (t *testGetter) getRecords(ctx context.Context) ([]*pbrc.Record, error) {
	if t.fail {
		return []*pbrc.Record{}, fmt.Errorf("Built to fail")
	}
	return t.records, nil
}

func (t *testGetter) updatePrice(ctx context.Context, instanceID, price int32) error {
	return nil
}

func getTestServer() *Server {
	s := Init()
	s.SkipLog = true
	s.GoServer.KSclient = *keystoreclient.GetTestClient(".test")

	return s
}

func TestSyncSales(t *testing.T) {
	s := getTestServer()
	s.getter = &testGetter{records: []*pbrc.Record{&pbrc.Record{Metadata: &pbrc.ReleaseMetadata{SaleId: 12}, Release: &pbgd.Release{InstanceId: 12}}}}

	s.syncSales(context.Background())

	found := false
	for _, sale := range s.config.Sales {
		if sale.InstanceId == 12 {
			found = true
		}
	}

	if !found {
		t.Errorf("Records were not synced: %v", s.config)
	}
}

func TestSyncSalesWithCacheHit(t *testing.T) {
	s := getTestServer()
	s.config.Sales = append(s.config.Sales, &pb.Sale{InstanceId: 12, LastUpdateTime: 12})
	s.getter = &testGetter{records: []*pbrc.Record{&pbrc.Record{Metadata: &pbrc.ReleaseMetadata{SaleId: 12}, Release: &pbgd.Release{InstanceId: 12}}}}

	s.syncSales(context.Background())

	found := false
	for _, sale := range s.config.Sales {
		if sale.InstanceId == 12 && sale.LastUpdateTime == 12 {
			found = true
		}
	}

	if !found {
		t.Errorf("Records were not synced: %v", s.config)
	}
}

func TestSyncSalesWithGetFail(t *testing.T) {
	s := getTestServer()
	s.getter = &testGetter{fail: true, records: []*pbrc.Record{&pbrc.Record{Metadata: &pbrc.ReleaseMetadata{SaleId: 12}, Release: &pbgd.Release{InstanceId: 12}}}}
	s.syncSales(context.Background())

	if len(s.config.Sales) > 0 {
		t.Errorf("Sales have synced somehow: %v", s.config)
	}
}

func TestUpateSales(t *testing.T) {
	s := getTestServer()
	s.config.Sales = append(s.config.Sales, &pb.Sale{InstanceId: 12, LastUpdateTime: 12})
	s.updateSales(context.Background())

	if s.config.Sales[0].LastUpdateTime == 12 {
		t.Errorf("This test needs updating: %v", s.config)
	}
}
