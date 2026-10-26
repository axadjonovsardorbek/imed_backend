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

type PromoCodeService struct {
	storage st.Storage
	rdb     *redis.Client
}

func NewPromoCodeService(storage *st.Storage, rdb *redis.Client) *PromoCodeService {
	return &PromoCodeService{storage: *storage, rdb: rdb}
}

func (s *PromoCodeService) getVersion() string {
	version := "1"
	if s.rdb != nil {
		if v, err := s.rdb.Get(context.Background(), "promocode:version").Result(); err == nil && v != "" {
			version = v
		}
	}
	return version
}

func (s *PromoCodeService) Create(ctx context.Context, req *m.PromoCodeCreateReq) (*m.ById, error) {
	res, err := s.storage.PromoCodeS.Create(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "promocode:version").Err()
	}

	return res, nil
}
func (s *PromoCodeService) GetById(ctx context.Context, req *m.ById) (*m.PromoCodeRes, error) {
	return s.storage.PromoCodeS.GetById(req)
}
func (s *PromoCodeService) GetAll(ctx context.Context, req *m.PromoCodeGetListReq) (*m.PaginationResponse, *m.PromoCodeList, error) {

	version := s.getVersion()

	key := fmt.Sprintf(
		"promocode:list:v%s_code:%s_type:%s_active:%s_page:%d_limit:%d",
		version,
		req.Code,
		req.DiscountType,
		req.IsActive,
		req.Filter.Page,
		req.Filter.Limit,
	)

	var res *m.PaginationResponse

	if s.rdb != nil {
		if b, err := s.rdb.Get(context.Background(), key).Bytes(); err == nil && len(b) > 0 {
			var cached m.PaginationResponse
			if err := json.Unmarshal(b, &cached); err == nil {
				return &cached, nil, nil
			}
		}
	}

	res, list, err := s.storage.PromoCodeS.GetAll(req)
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
func (s *PromoCodeService) Update(ctx context.Context, req *m.PromoCodeUpdateReq) (*m.Void, error) {
	res, err := s.storage.PromoCodeS.Update(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "promocode:version").Err()
	}

	return res, nil
}
func (s *PromoCodeService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	res, err := s.storage.PromoCodeS.Delete(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "promocode:version").Err()
	}

	return res, nil
}
func (s *PromoCodeService) Check(ctx context.Context, req *m.PromoCodeCheckReq) (*m.PromoCodeCheckRes, error) {
	return s.storage.PromoCodeS.Check(req)
}
func (s *PromoCodeService) GetByIdUserPromo(ctx context.Context, req *m.ById) (*m.PromoRedemptionRes, error) {
	return s.storage.PromoCodeS.GetByIdUserPromo(req)
}
func (s *PromoCodeService) GetAllUserPromo(ctx context.Context, req *m.PromoRedemptionGetListReq) (*m.PaginationResponse, *m.PromoRedemptionList, error) {

	version := s.getVersion()

	key := fmt.Sprintf(
		"promocode:user:v%s_promo:%s_user:%s_active:%s_page:%d_limit:%d",
		version,
		req.PromoId,
		req.UserId,
		req.IsActive,
		req.Filter.Page,
		req.Filter.Limit,
	)

	if s.rdb != nil {
		if b, err := s.rdb.Get(context.Background(), key).Bytes(); err == nil && len(b) > 0 {
			var cached m.PaginationResponse
			if err := json.Unmarshal(b, &cached); err == nil {
				return &cached, nil, nil
			}
		}
	}

	res, list, err := s.storage.PromoCodeS.GetAllUserPromo(req)
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
