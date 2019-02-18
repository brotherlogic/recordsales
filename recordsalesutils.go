package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordsales/proto"
)

func (s *Server) trimList(ctx context.Context, in []*pb.Sale) []*pb.Sale {
	// Trim out excess
	seen := make(map[int32]map[int32]bool)
	narch := []*pb.Sale{}
	for _, a := range in {
		add := false
		if _, ok := seen[a.InstanceId]; ok {
			if _, ok2 := seen[a.InstanceId][a.Price]; !ok2 {
				add = true
				seen[a.InstanceId][a.Price] = true
			}
		} else {
			seen[a.InstanceId] = make(map[int32]bool)
			seen[a.InstanceId][a.Price] = true
			add = true
		}

		if add {
			narch = append(narch, a)
		} else {
			s.RaiseIssue(ctx, "Trim Needed", "Need to trim archives", false)
		}
	}
	return narch
}

func (s *Server) syncSales(ctx context.Context) {
	records, err := s.getter.getRecords(ctx)

	if err != nil {
		s.Log(fmt.Sprintf("Get error: %v", err))
		return
	}

	for _, rec := range records {
		if rec.GetMetadata().SaleId > 0 {
			found := false
			for _, sale := range s.config.Sales {
				if sale.InstanceId == rec.GetRelease().InstanceId {
					found = true
					if !rec.GetMetadata().SaleDirty {
						oldSale := &pb.Sale{
							InstanceId:     sale.InstanceId,
							LastUpdateTime: sale.LastUpdateTime,
							Price:          sale.Price,
						}
						seen := false
						for _, arch := range s.config.Archives {
							if arch.InstanceId == oldSale.InstanceId && arch.Price == oldSale.Price {
								seen = true
							}
						}
						if !seen {
							s.config.Archives = append(s.config.Archives, oldSale)
						}
						sale.Price = rec.GetMetadata().SalePrice

					}
					break
				}
			}

			if !found {
				s.Log(fmt.Sprintf("Not found - adding %v", rec.GetRelease().Title))
				s.config.Sales = append(s.config.Sales, &pb.Sale{InstanceId: rec.GetRelease().InstanceId, LastUpdateTime: time.Now().Unix()})
			}

			//Remove record if it's sold
			if rec.GetMetadata().Category != pbrc.ReleaseMetadata_LISTED_TO_SELL {
				i := 0
				for i < len(s.config.Sales) {
					if s.config.Sales[i].InstanceId == rec.GetRelease().InstanceId {
						s.config.Sales = append(s.config.Sales[:i], s.config.Sales[i+1:]...)
					}
					i++
				}
			}

		}
	}

	s.save(ctx)
}

func (s *Server) updateSales(ctx context.Context) {
	s.updates++
	for _, sale := range s.config.Sales {
		if time.Now().Sub(time.Unix(sale.LastUpdateTime, 0)) > time.Hour*24*7 && sale.Price != 500 { //one week
			sale.LastUpdateTime = time.Now().Unix()
			newPrice := sale.Price - 500
			if newPrice < 500 {
				newPrice = 500
			}
			s.Log(fmt.Sprintf("Updating %v -> %v", sale.InstanceId, newPrice))
			s.getter.updatePrice(ctx, sale.InstanceId, newPrice)
		}
	}
	s.save(ctx)
}
