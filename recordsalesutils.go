package main

import (
	"fmt"
	"sort"
	"time"

	"golang.org/x/net/context"

	gdpb "github.com/brotherlogic/godiscogs"
	"github.com/brotherlogic/goserver/utils"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordsales/proto"
)

func (s *Server) isInPlay(ctx context.Context, r *pbrc.Record) bool {
	return false //s.testing
}

func (s *Server) runSales() {
	for true {
		t := s.runElection()

		//Wait between sale runs
		s.Log(fmt.Sprintf("Sleeping for %v", t))
		time.Sleep(t)
	}
}

func (s *Server) runElection() time.Duration {
	ecancel, err := s.Elect()
	defer ecancel()
	if err != nil {
		s.Log(fmt.Sprintf("Unable to elect: %v", err))
		return time.Minute * 5
	}

	ctx, cancel := utils.ManualContext("saleloop", "saleloop", time.Minute, true)
	config, err := s.load(ctx)
	cancel()
	if err != nil {
		s.Log(fmt.Sprintf("Unable to load config: %v", err))
		return time.Minute * 5
	}
	s.setOldest(config.GetSales())

	sort.SliceStable(config.Sales, func(i, j int) bool {
		return config.Sales[i].GetLastUpdateTime() < config.Sales[j].GetLastUpdateTime()
	})

	for i := range config.Sales {
		if time.Now().After(time.Unix(config.Sales[i].GetNextProcessTime(), 0)) {
			err = s.updateSales(config.Sales[i])
			config.Sales[i].NextProcessTime = time.Now().Add(time.Hour).Unix()

			ctx, cancel := utils.ManualContext("cs", "cs", time.Minute, true)
			s.save(ctx, config)
			cancel()

			//Next update time
			nut := time.Unix(config.Sales[i+1].GetLastUpdateTime(), 0).Add(time.Hour * 24 * 7)
			stime := nut.Sub(time.Now())
			if stime > 0 {
				return stime
			}
			return time.Minute * 5
		}
	}

	return time.Minute * 5
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
			s.RaiseIssue("Trim Needed", "Need to trim archives")
		}
	}
	return narch
}

func (s *Server) syncSales(ctx context.Context, iid int32) error {
	rec, err := s.getter.loadRecord(ctx, iid)
	if err != nil {
		return err
	}

	config, err := s.load(ctx)
	if err != nil {
		return err
	}

	found := false
	for _, sale := range config.Sales {
		if sale.InstanceId == rec.GetRelease().InstanceId {
			found = true
			if !rec.GetMetadata().SaleDirty {
				oldSale := &pb.Sale{
					InstanceId:     sale.InstanceId,
					LastUpdateTime: sale.LastUpdateTime,
					Price:          sale.Price,
				}
				seen := false
				for _, arch := range config.Archives {
					if arch.InstanceId == oldSale.InstanceId && arch.Price == oldSale.Price {
						seen = true
					}
				}
				if !seen {
					config.Archives = append(config.Archives, oldSale)
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

	if !found && rec.GetMetadata().GetSaleId() > 0 && (rec.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_LISTED_TO_SELL) {
		s.Log(fmt.Sprintf("NEW SALE: %v", rec.GetRelease().GetInstanceId()))
		config.Sales = append(config.Sales, &pb.Sale{InstanceId: rec.GetRelease().InstanceId, LastUpdateTime: time.Now().Unix()})
	}

	//Remove record if it's sold
	if found && rec.GetMetadata().Category != pbrc.ReleaseMetadata_LISTED_TO_SELL {
		s.Log(fmt.Sprintf("REMOVING %v -> %v, %v", rec.GetRelease().InstanceId, found, rec.GetMetadata().Category != pbrc.ReleaseMetadata_LISTED_TO_SELL && rec.GetMetadata().Category != pbrc.ReleaseMetadata_STALE_SALE))
		i := 0
		for i < len(config.Sales) {
			if config.Sales[i].InstanceId == rec.GetRelease().InstanceId {
				config.Sales = append(config.Sales[:i], config.Sales[i+1:]...)
			}
			i++
		}
	}

	return s.save(ctx, config)
}

func (s *Server) updateSales(sale *pb.Sale) error {
	cancel, err := s.ElectKey(fmt.Sprintf("%v", sale.GetInstanceId()))
	if err != nil {
		return err
	}
	defer cancel()

	time.Sleep(time.Second * 2)
	s.Log(fmt.Sprintf("Running update for %v", sale))
	time.Sleep(time.Second * 2)
	if !sale.OnHold {
		ctx, cancel := utils.ManualContext("updatesales", "updatesales", time.Minute, false)
		defer cancel()
		if time.Now().Sub(time.Unix(sale.LastUpdateTime, 0)) > time.Hour*24*7 && sale.Price > 499 { //one week
			sale.LastUpdateTime = time.Now().Unix()
			newPrice := sale.Price - 500
			if newPrice < 499 {
				newPrice = 499
			}
			err := s.getter.updatePrice(ctx, sale.InstanceId, newPrice)
			s.Log(fmt.Sprintf("Updating %v -> %v [%v]", sale.InstanceId, newPrice, err))
			s.getter.updateCategory(ctx, sale.InstanceId, pbrc.ReleaseMetadata_LISTED_TO_SELL)
			if err != nil {
				return err

			}
		} else if time.Now().Sub(time.Unix(sale.LastUpdateTime, 0)) > time.Hour*24*7*4 && sale.Price <= 499 && sale.Price > 200 { // one month
			s.Log(fmt.Sprintf("[%v] STALE for %v", sale.InstanceId, time.Now().Sub(time.Unix(sale.LastUpdateTime, 0))))
			s.getter.updateCategory(ctx, sale.InstanceId, pbrc.ReleaseMetadata_STALE_SALE)
			err := s.getter.updatePrice(ctx, sale.InstanceId, 200)
			if err != nil {
				return err
			}
		} else {
			/*We've done nothing to this entry - let's just trigger an update to refresh the cache
			s.Log(fmt.Sprintf("NO-OP UPDATE %v", sale))
			err := s.getter.updatePrice(ctx, sale.InstanceId, sale.Price)
			if err != nil {
				return err
			}*/
		}
	}
	return nil
}
