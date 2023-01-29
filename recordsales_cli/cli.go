package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/brotherlogic/goserver/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordsales/proto"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
)

func getRecord(ctx context.Context, instanceID int32) *pbrc.Record {
	host, port, err := utils.Resolve("recordcollection", "sales-cli")
	if err != nil {
		log.Fatalf("Unable to reach recordcollection: %v", err)
	}
	conn, err := grpc.Dial(host+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}

	client := pbrc.NewRecordCollectionServiceClient(conn)
	r, err := client.GetRecord(ctx, &pbrc.GetRecordRequest{InstanceId: instanceID})
	if err != nil {
		log.Fatalf("Unable to get records: %v", err)
	}

	return r.GetRecord()
}

func main() {
	ctx, cancel := utils.BuildContext("recordsales-cli-"+os.Args[1], "recordsales-cli")
	defer cancel()

	conn, err := utils.LFDialServer(ctx, "recordsales")
	defer conn.Close()

	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}

	client := pb.NewSaleServiceClient(conn)

	switch os.Args[1] {
	case "list":
		res, err := client.GetStale(ctx, &pb.GetStaleRequest{})
		if err != nil {
			log.Fatalf("Error on GET: %v", err)
		}
		fmt.Printf("Found %v stale sales\n", len(res.GetStaleSales()))
		for i, id := range res.GetStaleSales() {
			rec := getRecord(ctx, id.InstanceId)
			fmt.Printf("%v. %v\n", i, rec.GetRelease().Title)
		}
	case "all":
		res, err := client.GetSaleState(ctx, &pb.GetStateRequest{})
		if err != nil {
			log.Fatalf("Bad get: %v", err)
		}
		for i, sale := range res.GetSales() {
			fmt.Printf("%v. %v (%v)\n", i, sale.GetInstanceId(), sale.GetPrice())
		}
	case "get":
		val, _ := strconv.ParseInt(os.Args[2], 10, 32)
		res, err := client.GetSaleState(ctx, &pb.GetStateRequest{InstanceId: int32(val)})
		if err != nil {
			log.Fatalf("Cannot get: %v", err)
		}
		if len(res.GetSales()) == 0 {
			fmt.Printf("No sales found!\n")
		}
		for _, r := range res.GetSales() {
			fmt.Printf("%v - %v\n", time.Unix(r.GetLastUpdateTime(), 0), r.GetPrice())
		}
	case "sales":
		val, _ := strconv.ParseInt(os.Args[2], 10, 32)
		res, err := client.GetPrice(ctx, &pb.GetPriceRequest{Ids: []int32{int32(val)}})
		if err != nil {
			log.Fatalf("Cannot get: %v", err)
		}
		if len(res.GetPrices()[int32(val)].GetHistory()) == 0 {
			fmt.Printf("No sales found!\n")
		}
		for _, r := range res.GetPrices()[int32(val)].GetHistory() {
			fmt.Printf("%v - %v\n", time.Unix(r.GetDate(), 0), r.GetPrice())
		}
	case "force":
		val, _ := strconv.ParseInt(os.Args[2], 10, 32)
		client := pbrc.NewClientUpdateServiceClient(conn)
		resp, err := client.ClientUpdate(ctx, &pbrc.ClientUpdateRequest{InstanceId: int32(val)})

		fmt.Printf("%v and %v\n", resp, err)
	case "update":
		val, _ := strconv.ParseInt(os.Args[2], 10, 32)
		resp, err := client.UpdatePrice(ctx, &pb.UpdatePriceRequest{Id: int32(val)})

		fmt.Printf("%v and %v\n", resp, err)
	case "fullping":
		ctx2, cancel2 := utils.ManualContext("recordcollectioncli", time.Hour)
		defer cancel2()

		conn, err := utils.LFDialServer(ctx2, "recordsales")
		if err != nil {
			log.Fatalf("Argh: %v", err)
		}
		client := pbrc.NewClientUpdateServiceClient(conn)

		conn2, err := utils.LFDialServer(ctx2, "recordcollection")
		if err != nil {
			log.Fatalf("Cannot reach rc: %v", err)
		}
		defer conn2.Close()

		registry := pbrc.NewRecordCollectionServiceClient(conn2)
		ids, err := registry.QueryRecords(ctx2, &pbrc.QueryRecordsRequest{Query: &pbrc.QueryRecordsRequest_All{true}})
		if err != nil {
			log.Fatalf("Bad query: %v", err)
		}

		for i, id := range ids.GetInstanceIds() {
			fmt.Printf("PING %v -> %v\n", i, id)
			_, err = client.ClientUpdate(ctx2, &pbrc.ClientUpdateRequest{InstanceId: int32(id)})
			if err != nil {
				log.Fatalf("Error on GET: %v", err)
			}

		}
	default:
		fmt.Printf("Unknown command: %v\n", os.Args[1])
	}
}
