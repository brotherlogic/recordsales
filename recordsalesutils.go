package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	gdpb "github.com/brotherlogic/godiscogs/proto"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordsales/proto"
	"github.com/prometheus/client_golang/prometheus"
)

func (s *Server) metrics(config *pb.Config) {
	sales.Set(float64(len(config.GetSales())))

	for _, sale := range config.GetSales() {
		cost.With(prometheus.Labels{"id": fmt.Sprintf("%v", sale.GetInstanceId())}).Set(float64(sale.GetPrice()))
	}

	sizes.With(prometheus.Labels{"category": "sales"}).Set(float64(len(config.GetSales())))
	sizes.With(prometheus.Labels{"category": "archive"}).Set(float64(len(config.GetArchives())))
	sizes.With(prometheus.Labels{"category": "history"}).Set(float64(len(config.GetPriceHistory())))
}

func (s *Server) isInPlay(ctx context.Context, r *pbrc.Record) bool {
	return false //s.testing
}

func (s *Server) trimRecords(ctx context.Context, nrecs []*pbrc.Record) ([]*pbrc.Record, error) {
	recs := []*pbrc.Record{}

	for _, rec := range nrecs {
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

func (s *Server) syncSales(ctx context.Context, rec *pbrc.Record) error {

	config, err := s.load(ctx)
	if err != nil {
		return err
	}

	found := false
	for _, sale := range config.Sales {
		if sale.InstanceId == rec.GetRelease().InstanceId {
			found = true

			if rec.GetMetadata().GetSaleBudget() == "" {
				rec.GetMetadata().SaleBudget = "float2024"
			}

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
				s.CtxLog(ctx, fmt.Sprintf("NEWPRICE %v -> %v, %v", rec.GetRelease().GetInstanceId(), rec.GetMetadata().GetSalePrice(), rec.GetMetadata().GetCurrentSalePrice()))
				sale.Price = rec.GetMetadata().CurrentSalePrice
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
		s.CtxLog(ctx, fmt.Sprintf("NEWSALE: %v", rec.GetRelease().GetInstanceId()))
		config.Sales = append(config.Sales, &pb.Sale{InstanceId: rec.GetRelease().InstanceId, LastUpdateTime: time.Now().Unix()})
	}

	//Remove record if it's sold
	if found && rec.GetMetadata().Category != pbrc.ReleaseMetadata_LISTED_TO_SELL {
		s.CtxLog(ctx, fmt.Sprintf("REMSALE: %v -> %v", rec.GetRelease().GetInstanceId(), rec.GetMetadata().GetCategory()))
		i := 0
		for i < len(config.Sales) {
			if config.Sales[i].InstanceId == rec.GetRelease().InstanceId {
				config.Sales = append(config.Sales[:i], config.Sales[i+1:]...)
			}
			i++
		}
	}

	// Remove record if it is NOT_FOR_SALE or EXPIRED
	var nsales []*pb.Sale
	for _, sale := range config.GetSales() {
		if sale.GetInstanceId() == rec.GetRelease().GetInstanceId() {
			if rec.GetMetadata().GetSaleState() == gdpb.SaleState_FOR_SALE {
				nsales = append(nsales, sale)
			} else {
				s.CtxLog(ctx, fmt.Sprintf("REMOVING %v -> %v", rec.GetRelease().GetInstanceId(), rec.GetMetadata().GetBoxState()))
			}
		} else {
			nsales = append(nsales, sale)
		}
	}

	config.Sales = nsales

	return s.save(ctx, config)
}
