package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	gdpb "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordsales/proto"
)

func (s *Server) isInPlay(ctx context.Context, r *pbrc.Record) bool {
	return s.testing
}

func (s *Server) trimRecords(ctx context.Context, nrecs []*pbrc.Record) ([]*pbrc.Record, error) {
	recs := []*pbrc.Record{}

	for _, rec := range nrecs {
		s.Log(fmt.Sprintf("%v is in play? %v", rec.GetRelease().GetInstanceId(), s.isInPlay(ctx, rec)))
		if s.isInPlay(ctx, rec) {
			//Ensure the record is for sale if it needs to be
			if rec.GetMetadata().SaleState == gdpb.SaleState_EXPIRED {
				err := s.getter.updatePrice(ctx, rec.GetRelease().GetInstanceId(), rec.GetMetadata().GetSalePrice())
				if err != nil {
					return recs, err
				}
			}

			recs = append(recs, rec)
		} else if rec.GetMetadata().SaleState != gdpb.SaleState_EXPIRED && !rec.GetMetadata().GetSaleDirty() {
			err := s.getter.expireSale(ctx, rec.GetRelease().GetInstanceId())
			if err != nil {
				return recs, err
			}
		}
	}

	return recs, nil
}

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

func (s *Server) syncSales(ctx context.Context) error {
	nrecords, err := s.getter.getListedRecords(ctx)
	if err != nil {
		return err
	}

	records, err := s.trimRecords(ctx, nrecords)
	if err != nil {
		return err
	}

	s.Log(fmt.Sprintf("Running on %v records", len(records)))

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
							s.Log(fmt.Sprintf("NEW SALE: %v", oldSale))
							s.config.Archives = append(s.config.Archives, oldSale)
						}
						sale.Price = rec.GetMetadata().SalePrice
						if sale.Price == 0 {
							sale.Price = rec.GetMetadata().GetCurrentSalePrice()
						}
						sale.LastUpdateTime = rec.GetMetadata().LastSalePriceUpdate
						sale.OnHold = rec.GetMetadata().GetSaleState() == gdpb.SaleState_EXPIRED

					}
					break

				}
			}

			if !found {
				s.config.Sales = append(s.config.Sales, &pb.Sale{InstanceId: rec.GetRelease().InstanceId, LastUpdateTime: time.Now().Unix()})
			}

			//Remove record if it's sold
			if rec.GetMetadata().Category != pbrc.ReleaseMetadata_LISTED_TO_SELL && rec.GetMetadata().Category != pbrc.ReleaseMetadata_STALE_SALE {
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
	return nil
}

func (s *Server) updateSales(ctx context.Context) error {
	s.updates++
	for _, sale := range s.config.Sales {
		if !sale.OnHold {
			if time.Now().Sub(time.Unix(sale.LastUpdateTime, 0)) > time.Hour*24*7*8 && sale.Price != 499 && sale.Price != 200 { //two months
				sale.LastUpdateTime = time.Now().Unix()
				newPrice := sale.Price - 500
				if newPrice < 499 {
					newPrice = 499
				}
				s.Log(fmt.Sprintf("Updating %v -> %v", sale.InstanceId, newPrice))
				err := s.getter.updatePrice(ctx, sale.InstanceId, newPrice)
				s.getter.updateCategory(ctx, sale.InstanceId, pbrc.ReleaseMetadata_LISTED_TO_SELL)
				if err != nil {
					return err
				}
			} else if time.Now().Sub(time.Unix(sale.LastUpdateTime, 0)) > time.Hour*24*7*4 && (sale.Price == 499 || sale.Price == 498) { // one month
				s.Log(fmt.Sprintf("[%v] STALE for %v", sale.InstanceId, time.Now().Sub(time.Unix(sale.LastUpdateTime, 0))))
				s.getter.updateCategory(ctx, sale.InstanceId, pbrc.ReleaseMetadata_STALE_SALE)
				err := s.getter.updatePrice(ctx, sale.InstanceId, 200)
				if err != nil {
					return err
				}
			}
		}
	}
	s.config.LastSaleRun = time.Now().Unix()
	s.save(ctx)
	return nil
}
