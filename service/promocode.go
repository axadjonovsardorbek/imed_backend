package service

import (
	"context"
	"encoding/json"
	"fmt"
	m "imed/pkg/models"
	st "imed/storage/postgres"
	"time"

	"github.com/go-redis/redis"
)

type PromoCodeService struct {
	storage st.Storage
	rdb     *redis.Client
}

func NewPromoCodeService(storage *st.Storage, rdb *redis.Client) *PromoCodeService {
	return &PromoCodeService{storage: *storage, rdb: rdb}
}

func (s *PromoCodeService) Create(ctx context.Context, req *m.PromoCodeCreateReq) (*m.ById, error) {
	return s.storage.PromoCodeS.Create(req)
}
func (s *PromoCodeService) GetById(ctx context.Context, req *m.ById) (*m.PromoCodeRes, error) {
	return s.storage.PromoCodeS.GetById(req)
}
func (s *PromoCodeService) GetAll(ctx context.Context, req *m.PromoCodeGetListReq) (*m.PaginationResponse, *m.PromoCodeList, error) {
	res := &m.PaginationResponse{}

	key := fmt.Sprintf("code:%s_discount_type:%s_is_active:%s_page:%d_limit:%d",
		req.Code, req.DiscountType, req.IsActive, req.Filter.Page, req.Filter.Limit)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res)
		}
	}

	if res.Total == 0 {
		res, _, err := s.storage.PromoCodeS.GetAll(req)
		if err != nil {
			return nil, nil, err
		}

		if s.rdb != nil {
			if b, err := json.Marshal(res); err == nil {
				_ = s.rdb.Set(key, b, 60*time.Second).Err()
			}
		}
	}

	return res, nil, nil
}
func (s *PromoCodeService) Update(ctx context.Context, req *m.PromoCodeUpdateReq) (*m.Void, error) {
	return s.storage.PromoCodeS.Update(req)
}
func (s *PromoCodeService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.PromoCodeS.Delete(req)
}
func (s *PromoCodeService) Check(ctx context.Context, req *m.PromoCodeCheckReq) (*m.PromoCodeCheckRes, error) {
	return s.storage.PromoCodeS.Check(req)
}
