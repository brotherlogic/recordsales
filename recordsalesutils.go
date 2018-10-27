package main

import (
	"time"

	"golang.org/x/net/context"

	pb "github.com/brotherlogic/recordsales/proto"
)

func (s *Server) syncSales(ctx context.Context) {
	records, err := s.getter.getRecords(ctx)

	if err != nil {
		return
	}

	for _, rec := range records {
		if rec.GetMetadata().SaleId > 0 {
			found := false
			for _, s := range s.config.Sales {
				if s.InstanceId == rec.GetRelease().InstanceId {
					found = true
					break
				}
			}

			if !found {
				s.config.Sales = append(s.config.Sales, &pb.Sale{InstanceId: rec.GetRelease().InstanceId, LastUpdateTime: time.Now().Unix()})
			}
		}
	}
}
