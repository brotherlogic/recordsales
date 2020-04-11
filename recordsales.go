package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/keystore/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pbgd "github.com/brotherlogic/godiscogs"
	pbg "github.com/brotherlogic/goserver/proto"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordsales/proto"
)

type getter interface {
	getListedRecords(ctx context.Context) ([]*pbrc.Record, error)
	updatePrice(ctx context.Context, instanceID, price int32) error
	updateCategory(ctx context.Context, instanceID int32, category pbrc.ReleaseMetadata_Category)
	expireSale(ctx context.Context, instanceID int32) error
}

type prodGetter struct {
	dial func(server string) (*grpc.ClientConn, error)
}

func (p *prodGetter) getListedRecords(ctx context.Context) ([]*pbrc.Record, error) {
	conn, err := p.dial("recordcollection")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := pbrc.NewRecordCollectionServiceClient(conn)

	r, err := client.QueryRecords(ctx, &pbrc.QueryRecordsRequest{Query: &pbrc.QueryRecordsRequest_Category{pbrc.ReleaseMetadata_LISTED_TO_SELL}})
	if err != nil {
		return nil, err
	}
	instanceIds := r.GetInstanceIds()
	r, err = client.QueryRecords(ctx, &pbrc.QueryRecordsRequest{Query: &pbrc.QueryRecordsRequest_Category{pbrc.ReleaseMetadata_STALE_SALE}})
	if err != nil {
		return nil, err
	}
	instanceIds = append(instanceIds, r.GetInstanceIds()...)

	records := []*pbrc.Record{}
	for _, id := range instanceIds {
		r, err := client.GetRecord(ctx, &pbrc.GetRecordRequest{InstanceId: id})
		if err != nil {
			return nil, err
		}

		records = append(records, r.GetRecord())
	}

	return records, nil
}

func (p *prodGetter) updateCategory(ctx context.Context, instanceID int32, category pbrc.ReleaseMetadata_Category) {
	conn, err := p.dial("recordcollection")
	if err != nil {
		return
	}
	defer conn.Close()

	client := pbrc.NewRecordCollectionServiceClient(conn)
	update := &pbrc.UpdateRecordRequest{Update: &pbrc.Record{Release: &pbgd.Release{InstanceId: instanceID}, Metadata: &pbrc.ReleaseMetadata{Category: category}}}
	client.UpdateRecord(ctx, update)
}

func (p *prodGetter) updatePrice(ctx context.Context, instanceID, price int32) error {
	conn, err := p.dial("recordcollection")
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pbrc.NewRecordCollectionServiceClient(conn)
	update := &pbrc.UpdateRecordRequest{Update: &pbrc.Record{Release: &pbgd.Release{InstanceId: instanceID}, Metadata: &pbrc.ReleaseMetadata{NewSalePrice: price}}}
	_, err = client.UpdateRecord(ctx, update)
	return err
}

func (p *prodGetter) expireSale(ctx context.Context, instanceID int32) error {
	conn, err := p.dial("recordcollection")
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pbrc.NewRecordCollectionServiceClient(conn)
	update := &pbrc.UpdateRecordRequest{Update: &pbrc.Record{Release: &pbgd.Release{InstanceId: instanceID}, Metadata: &pbrc.ReleaseMetadata{ExpireSale: true}}}
	_, err = client.UpdateRecord(ctx, update)
	return err
}

const (
	// KEY - where we store sale info
	KEY = "/github.com/brotherlogic/recordsales/config"
)

//Server main server type
type Server struct {
	*goserver.GoServer
	config  *pb.Config
	getter  getter
	updates int64
	testing bool
}

// Init builds the server
func Init() *Server {
	s := &Server{
		&goserver.GoServer{},
		&pb.Config{},
		&prodGetter{},
		int64(0),
		true,
	}
	s.getter = &prodGetter{s.DialMaster}
	return s
}

func (s *Server) save(ctx context.Context) {
	s.KSclient.Save(ctx, KEY, s.config)
}

func (s *Server) load(ctx context.Context) error {
	config := &pb.Config{}
	data, _, err := s.KSclient.Read(ctx, KEY, config)

	if err != nil {
		return err
	}

	s.config = data.(*pb.Config)

	s.config.Archives = s.trimList(ctx, s.config.Archives)
	return nil
}

// DoRegister does RPC registration
func (s *Server) DoRegister(server *grpc.Server) {
	pb.RegisterSaleServiceServer(server, s)
}

// ReportHealth alerts if we're not healthy
func (s *Server) ReportHealth() bool {
	return true
}

// Shutdown the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.save(ctx)
	return nil
}

// Mote promotes/demotes this server
func (s *Server) Mote(ctx context.Context, master bool) error {
	if master {
		err := s.load(ctx)
		return err
	}

	return nil
}

// GetState gets the state of the server
func (s *Server) GetState() []*pbg.State {
	vals := ""
	for _, a := range s.config.Archives {
		if a.InstanceId == 19867545 {
			vals += fmt.Sprintf("%v [%v],", a.Price, time.Unix(a.LastUpdateTime, 0))
		}
	}
	sum := int32(0)
	pr := int32(0)
	oldest := time.Now().Unix()
	for _, s := range s.config.Sales {
		if s.LastUpdateTime < oldest {
			oldest = s.LastUpdateTime
		}
		sum += s.Price
		if s.InstanceId == 330510403 {
			pr += s.Price
		}
	}

	return []*pbg.State{
		&pbg.State{Key: "oldest", TimeValue: oldest},
		&pbg.State{Key: "last_sale_run", TimeValue: s.config.LastSaleRun},
		&pbg.State{Key: "active_sales", Value: int64(len(s.config.Sales))},
		&pbg.State{Key: "archive_sales", Value: int64(len(s.config.Archives))},
		&pbg.State{Key: "updates", Value: s.updates},
		&pbg.State{Key: "sum_sales", Value: int64(sum)},
		&pbg.State{Key: "tracker", Text: vals},
		&pbg.State{Key: "test", Text: "testing123"},
		&pbg.State{Key: "trac", Value: int64(pr)},
	}
}

func (s *Server) checkSaleTime(ctx context.Context) error {
	if time.Now().Sub(time.Unix(s.config.LastSaleRun, 0)) > time.Hour*24*7 {
		s.RaiseIssue(ctx, "Sale Problem", fmt.Sprintf("Last sale run was %v", time.Unix(s.config.LastSaleRun, 0)), false)
	}
	return nil
}

func main() {
	var quiet = flag.Bool("quiet", false, "Show all output")
	flag.Parse()

	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}
	server := Init()
	server.GoServer.KSclient = *keystoreclient.GetClient(server.DialMaster)
	server.PrepServer()
	server.Register = server

	err := server.RegisterServerV2("recordsales", false, false)
	if err != nil {
		return
	}

	server.RegisterRepeatingTask(server.syncSales, "sync_sales", time.Minute*5)
	server.RegisterRepeatingTask(server.checkSaleTime, "check_sale_time", time.Hour)
	server.RegisterRepeatingTask(server.updateSales, "update_sales", time.Minute*5)

	fmt.Printf("%v", server.Serve())
}
