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

func TestTrim(t *testing.T) {
	s := getTestServer()
	nlist := s.trimList(context.Background(), []*pb.Sale{&pb.Sale{InstanceId: 123, Price: 2020}, &pb.Sale{InstanceId: 124, Price: 20}, &pb.Sale{InstanceId: 124, Price: 20}, &pb.Sale{InstanceId: 123, Price: 20}})
	if len(nlist) != 3 {
		t.Errorf("Trim Error: %v", nlist)
	}
}

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

func (t *testGetter) updateCategory(ctx context.Context, instanceID int32, category pbrc.ReleaseMetadata_Category) {
}

func getTestServer() *Server {
	s := Init()
	s.SkipLog = true
	s.GoServer.KSclient = *keystoreclient.GetTestClient(".test")

	return s
}

func TestSyncSales(t *testing.T) {
	s := getTestServer()
	s.getter = &testGetter{records: []*pbrc.Record{&pbrc.Record{Metadata: &pbrc.ReleaseMetadata{SaleId: 12, Category: pbrc.ReleaseMetadata_LISTED_TO_SELL}, Release: &pbgd.Release{InstanceId: 12}}}}

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
	s.getter = &testGetter{records: []*pbrc.Record{&pbrc.Record{Metadata: &pbrc.ReleaseMetadata{SaleId: 12, Category: pbrc.ReleaseMetadata_LISTED_TO_SELL}, Release: &pbgd.Release{InstanceId: 12}}}}

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

func TestSyncSalesWithArchive(t *testing.T) {
	s := getTestServer()
	s.config.Sales = append(s.config.Sales, &pb.Sale{InstanceId: 12, LastUpdateTime: 12, Price: 200})
	s.config.Archives = append(s.config.Archives, &pb.Sale{InstanceId: 12, Price: 200})
	s.getter = &testGetter{records: []*pbrc.Record{&pbrc.Record{Metadata: &pbrc.ReleaseMetadata{SaleId: 12, SalePrice: 200}, Release: &pbgd.Release{InstanceId: 12}}}}

	s.syncSales(context.Background())

	if len(s.config.Archives) != 1 {
		t.Errorf("Too much archive: %v", s.config.Archives)
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

func TestRemoveRecordOnceSold(t *testing.T) {
	s := getTestServer()
	s.getter = &testGetter{records: []*pbrc.Record{&pbrc.Record{Release: &pbgd.Release{InstanceId: 1}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_SOLD_ARCHIVE, SaleId: 12345}}}}
	s.config.Sales = append(s.config.Sales, &pb.Sale{InstanceId: 1})

	s.syncSales(context.Background())

	if len(s.config.Sales) != 0 && len(s.config.Archives) != 1 {
		t.Errorf("Record sold has not been removed and added to archive")
	}
}
