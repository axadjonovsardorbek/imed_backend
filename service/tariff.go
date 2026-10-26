package service

import (
	"context"
	"encoding/json"
	"fmt"
	m "imed/pkg/models"
	st "imed/storage/postgres"
	"time"

	"github.com/redis/go-redis/v9"
)

type TariffService struct {
	storage st.Storage
	rdb     *redis.Client
}

func NewTariffService(storage *st.Storage, rdb *redis.Client) *TariffService {
	return &TariffService{storage: *storage, rdb: rdb}
}

func (s *TariffService) getVersion() string {
	version := "1"
	if s.rdb != nil {
		if v, err := s.rdb.Get(context.Background(), "tariff:version").Result(); err == nil && v != "" {
			version = v
		}
	}
	return version
}

func (s *TariffService) Create(ctx context.Context, req *m.TariffCreateReq) (*m.ById, error) {
	res, err := s.storage.TariffS.Create(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "tariff:version").Err()
	}

	return res, nil
}

func (s *TariffService) GetById(ctx context.Context, req *m.ById) (*m.TariffRes, error) {
	return s.storage.TariffS.GetById(req)
}

func (s *TariffService) GetAll(ctx context.Context, req *m.TariffGetListReq) (*m.PaginationResponse, *m.TariffList, error) {

	version := s.getVersion()

	key := fmt.Sprintf(
		"tariff:list:v%s_duration:%d",
		version,
		req.Duration,
	)

	if s.rdb != nil {
		if b, err := s.rdb.Get(context.Background(), key).Bytes(); err == nil && len(b) > 0 {
			var cached m.PaginationResponse
			if err := json.Unmarshal(b, &cached); err == nil {
				return &cached, nil, nil
			}
		}
	}

	res, list, err := s.storage.TariffS.GetAll(req)
	if err != nil {
		return nil, nil, err
	}

	if s.rdb != nil {
		if b, err := json.Marshal(res); err == nil {
			_ = s.rdb.Set(context.Background(), key, b, 5*time.Minute).Err()
		}
	}

	return res, list, nil
}

func (s *TariffService) Update(ctx context.Context, req *m.TariffUpdateReq) (*m.Void, error) {
	res, err := s.storage.TariffS.Update(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "tariff:version").Err()
	}

	return res, nil
}

func (s *TariffService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	res, err := s.storage.TariffS.Delete(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "tariff:version").Err()
	}

	return res, nil
}
