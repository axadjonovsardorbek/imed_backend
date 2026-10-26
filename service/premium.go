package service

import (
	"context"
	"fmt"

	m "imed/pkg/models"
	st "imed/storage/postgres"

	"github.com/redis/go-redis/v9"
)

type PremiumService struct {
	storage st.Storage
	rdb     *redis.Client
}

func NewPremiumService(storage *st.Storage, rdb *redis.Client) *PremiumService {
	return &PremiumService{storage: *storage, rdb: rdb}
}

func (s *PremiumService) GetSettings(ctx context.Context) (*m.PremiumSettings, error) {
	return s.storage.PremiumS.GetSettings(ctx)
}

func (s *PremiumService) UpdateSettings(ctx context.Context, req *m.PremiumSettings) (*m.Void, error) {
	if req.Price < 0 {
		return nil, fmt.Errorf("price must not be negative")
	}
	if req.DurationMonths <= 0 {
		req.DurationMonths = 1
	}
	return s.storage.PremiumS.UpdateSettings(ctx, req)
}

func (s *PremiumService) GetInfo(ctx context.Context, userId string) (*m.PremiumInfoRes, error) {
	settings, err := s.storage.PremiumS.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	status, err := s.storage.PremiumS.GetStatus(ctx, userId)
	if err != nil {
		return nil, err
	}
	return &m.PremiumInfoRes{PremiumSettings: *settings, AccessStatus: *status}, nil
}

func (s *PremiumService) CreateOrder(ctx context.Context, req *m.PremiumOrderReq) (*m.OrderCreateRes, error) {
	if req.PromoCodeId == "string" {
		req.PromoCodeId = ""
	}
	res, err := s.storage.PremiumS.CreateOrder(ctx, req)
	if err != nil {
		return nil, err
	}
	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "order:version").Err()
	}
	return res, nil
}

func (s *PremiumService) SetUstoz(ctx context.Context, req *m.UstozReq) (*m.Void, error) {
	return s.storage.PremiumS.SetUstoz(ctx, req)
}
