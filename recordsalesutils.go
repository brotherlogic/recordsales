package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	pb "github.com/brotherlogic/recordsales/proto"
)

func (s *Server) syncSales(ctx context.Context) {
	records, err := s.getter.getRecords(ctx)

	if err != nil {
		s.Log(fmt.Sprintf("Get error: %v", err))
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
				s.Log(fmt.Sprintf("Not found - adding %v", rec.GetRelease().Title))
				s.config.Sales = append(s.config.Sales, &pb.Sale{InstanceId: rec.GetRelease().InstanceId, LastUpdateTime: time.Now().Unix()})
			}
		}
	}

	s.save(ctx)
}

func (s *Server) updateSales(ctx context.Context) {
	s.updates++
	for _, sale := range s.config.Sales {
		s.Log(fmt.Sprintf("DELAY %v", time.Now().Sub(time.Unix(sale.LastUpdateTime, 0))))
		if time.Now().Sub(time.Unix(sale.LastUpdateTime, 0)) > time.Minute*30 { //one week
			sale.LastUpdateTime = time.Now().Unix()
			s.RaiseIssue(ctx, "Updating Sale Price", fmt.Sprintf("Updating price of %v (Currently %v)", sale.InstanceId, sale.Price), false)
		}
	}
	s.save(ctx)
}
