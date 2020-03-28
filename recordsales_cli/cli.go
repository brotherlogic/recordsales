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
	host, port, err := utils.Resolve("recordsales", "sales-cli")
	if err != nil {
		log.Fatalf("Unable to reach sales: %v", err)
	}
	conn, err := grpc.Dial(host+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}

	client := pb.NewSaleServiceClient(conn)
	ctx, cancel := utils.BuildContext("recordsales-cli-"+os.Args[1], "recordsales-cli")
	defer cancel()

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
	case "get":
		val, _ := strconv.Atoi(os.Args[2])
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
	default:
		fmt.Sprintf("Unknown command\n")
	}
}
