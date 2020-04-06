package main

import (
	"fmt"
	"testing"
	"time"

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
	records    []*pbrc.Record
	fail       bool
	failExpire bool
}

func (t *testGetter) getListedRecords(ctx context.Context) ([]*pbrc.Record, error) {
	if t.fail {
		return []*pbrc.Record{}, fmt.Errorf("Built to fail")
	}
	return t.records, nil
}

func (t *testGetter) updatePrice(ctx context.Context, instanceID, price int32) error {
	if t.fail {
		return fmt.Errorf("Built to fail")
	}
	return nil
}

func (t *testGetter) expireSale(ctx context.Context, price int32) error {
	if t.failExpire {
		return fmt.Errorf("Built to fail")
	}
	return nil
}

func (t *testGetter) updateCategory(ctx context.Context, instanceID int32, category pbrc.ReleaseMetadata_Category) {
}

func getTestServer() *Server {
	s := Init()
	s.SkipLog = true
	s.SkipIssue = true
	s.GoServer.KSclient = *keystoreclient.GetTestClient(".test")
	s.getter = &testGetter{}
	s.testing = false

	return s
}

func TestSyncSales(t *testing.T) {
	s := getTestServer()
	s.testing = true
	s.getter = &testGetter{records: []*pbrc.Record{&pbrc.Record{Metadata: &pbrc.ReleaseMetadata{SaleId: 12, Category: pbrc.ReleaseMetadata_LISTED_TO_SELL}, Release: &pbgd.Release{InstanceId: 177077893}}}}

	s.syncSales(context.Background())

	found := false
	for _, sale := range s.config.Sales {
		if sale.InstanceId == 177077893 {
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
	s.getter = &testGetter{records: []*pbrc.Record{&pbrc.Record{Metadata: &pbrc.ReleaseMetadata{SaleId: 12, Category: pbrc.ReleaseMetadata_LISTED_TO_SELL, LastSalePriceUpdate: 12}, Release: &pbgd.Release{InstanceId: 12}}}}

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
	s.testing = true
	s.config.Sales = append(s.config.Sales, &pb.Sale{InstanceId: 177077893, LastUpdateTime: 12, Price: 200})
	s.config.Archives = append(s.config.Archives, &pb.Sale{InstanceId: 177077893, Price: 200})
	s.getter = &testGetter{records: []*pbrc.Record{&pbrc.Record{Metadata: &pbrc.ReleaseMetadata{SaleId: 12, SalePrice: 200}, Release: &pbgd.Release{InstanceId: 177077893}}}}

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

func TestSyncSalesWithExpireFail(t *testing.T) {
	s := getTestServer()
	s.testing = false
	s.getter = &testGetter{failExpire: true, records: []*pbrc.Record{&pbrc.Record{Metadata: &pbrc.ReleaseMetadata{SaleId: 12}, Release: &pbgd.Release{InstanceId: 12, Formats: []*pbgd.Format{&pbgd.Format{Descriptions: []string{"7"}}}}}}}
	s.syncSales(context.Background())

	if len(s.config.Sales) > 0 {
		t.Errorf("Sales have synced somehow: %v", s.config)
	}
}

func TestUpdateSalesWithFail(t *testing.T) {
	s := getTestServer()
	s.getter = &testGetter{fail: true}
	s.config.Sales = append(s.config.Sales, &pb.Sale{InstanceId: 177077893, LastUpdateTime: 12})
	err := s.updateSales(context.Background())
	if err == nil {
		t.Errorf("Update did not fail")
	}
}

func TestUpdateSales(t *testing.T) {
	s := getTestServer()
	s.config.Sales = append(s.config.Sales, &pb.Sale{InstanceId: 177077893, LastUpdateTime: 12})
	err := s.updateSales(context.Background())
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	if s.config.Sales[0].LastUpdateTime == 12 {
		t.Errorf("This test needs updating: %v", s.config)
	}

	if time.Now().Sub(time.Unix(s.config.LastSaleRun, 0)) > time.Minute {
		t.Errorf("Time has not been updated: %v", time.Unix(s.config.LastSaleRun, 0))
	}
}

func TestUpdateSalesWhenOnHold(t *testing.T) {
	s := getTestServer()
	s.config.Sales = append(s.config.Sales, &pb.Sale{InstanceId: 177077893, LastUpdateTime: 12, OnHold: true})
	err := s.updateSales(context.Background())
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	if s.config.Sales[0].LastUpdateTime != 12 {
		t.Errorf("On Hold sale was updated")
	}
}

func TestUpdateSalesWithStale(t *testing.T) {
	s := getTestServer()
	s.config.Sales = append(s.config.Sales, &pb.Sale{InstanceId: 12, LastUpdateTime: 12, Price: 499})
	s.updateSales(context.Background())

	if s.config.Sales[0].LastUpdateTime != 12 {
		t.Errorf("This test needs updating: %v", s.config)
	}
}

func TestUpdateSalesWithStaleFail(t *testing.T) {
	s := getTestServer()
	s.getter = &testGetter{fail: true}
	s.config.Sales = append(s.config.Sales, &pb.Sale{InstanceId: 177077893, LastUpdateTime: 12, Price: 499})
	err := s.updateSales(context.Background())

	if err == nil {
		t.Errorf("Test did not fail")
	}

	if s.config.Sales[0].LastUpdateTime != 12 {
		t.Errorf("This test needs updating: %v", s.config)
	}
}

func TestRemoveRecordOnceSold(t *testing.T) {
	s := getTestServer()
	s.testing = true
	s.getter = &testGetter{records: []*pbrc.Record{&pbrc.Record{Release: &pbgd.Release{InstanceId: 177077893}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_SOLD_ARCHIVE, SaleId: 12345}}}}
	s.config.Sales = append(s.config.Sales, &pb.Sale{InstanceId: 177077893})

	s.syncSales(context.Background())

	if len(s.config.Sales) != 0 && len(s.config.Archives) != 1 {
		t.Errorf("Record sold has not been removed and added to archive")
	}
}

func TestNotInPlay(t *testing.T) {
	s := getTestServer()
	if s.isInPlay(context.Background(), &pbrc.Record{Release: &pbgd.Release{Formats: []*pbgd.Format{&pbgd.Format{Descriptions: []string{"7\""}}}}, Metadata: &pbrc.ReleaseMetadata{}}) {
		t.Errorf("All records are not in play")
	}
}

func TestInPlay(t *testing.T) {
	s := getTestServer()
	s.testing = true
	if !s.isInPlay(context.Background(), &pbrc.Record{Release: &pbgd.Release{Formats: []*pbgd.Format{&pbgd.Format{Descriptions: []string{"12\""}}}}, Metadata: &pbrc.ReleaseMetadata{}}) {
		t.Errorf("All records are not in play")
	}
}

func TestSaleTrim(t *testing.T) {
	s := getTestServer()
	s.testing = true
	recs, err := s.trimRecords(context.Background(), []*pbrc.Record{
		&pbrc.Record{Release: &pbgd.Release{InstanceId: 356769827, Formats: []*pbgd.Format{&pbgd.Format{Descriptions: []string{"7\""}}}}, Metadata: &pbrc.ReleaseMetadata{}},
		&pbrc.Record{Release: &pbgd.Release{InstanceId: 2}, Metadata: &pbrc.ReleaseMetadata{}},
	})

	if err != nil {
		t.Fatalf("Error in trim: %v", err)
	}

	if len(recs) != 2 {
		t.Errorf("Records were not trimmed: %v", len(recs))
	}
}
func TestSaleTrimRestoreFail(t *testing.T) {
	s := getTestServer()
	s.getter = &testGetter{fail: true}
	s.testing = true
	tr, err := s.trimRecords(context.Background(), []*pbrc.Record{
		&pbrc.Record{Release: &pbgd.Release{InstanceId: 356769827, Formats: []*pbgd.Format{&pbgd.Format{Descriptions: []string{"7\""}}}}, Metadata: &pbrc.ReleaseMetadata{}},
		&pbrc.Record{Release: &pbgd.Release{InstanceId: 2}, Metadata: &pbrc.ReleaseMetadata{SaleState: pbgd.SaleState_EXPIRED}},
	})

	if err == nil {
		t.Fatalf("Error in trim: %v", tr)
	}
}

func TestSaleTrimFail(t *testing.T) {
	s := getTestServer()
	s.testing = false
	s.getter = &testGetter{failExpire: true}
	_, err := s.trimRecords(context.Background(), []*pbrc.Record{
		&pbrc.Record{Release: &pbgd.Release{InstanceId: 356769827, Formats: []*pbgd.Format{&pbgd.Format{Descriptions: []string{"7\""}}}}, Metadata: &pbrc.ReleaseMetadata{}},
		&pbrc.Record{Release: &pbgd.Release{InstanceId: 2}, Metadata: &pbrc.ReleaseMetadata{SaleState: pbgd.SaleState_EXPIRED}},
	})

	if err == nil {
		t.Fatalf("Trim did not fail")
	}

}
