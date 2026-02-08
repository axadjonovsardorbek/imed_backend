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

type OrderService struct {
	storage st.Storage
	rdb     *redis.Client
}

func NewOrderService(storage *st.Storage, rdb *redis.Client) *OrderService {
	return &OrderService{storage: *storage, rdb: rdb}
}

func (s *OrderService) Create(ctx context.Context, req *m.OrderCreateReq) (*m.OrderCreateRes, error) {
	if req.PromoCodeId == "" || req.PromoCodeId == "string" {
		res, err := s.storage.OrderS.Create(req)
		if err != nil {
			return nil, err
		}
		fmt.Println("Order created without promo code:", res)
		return res, nil
	} else {
		res, err := s.storage.OrderS.CreateWithPromoCode(req)
		if err != nil {
			return nil, err
		}

		fmt.Println("Applying promo code:", req.PromoCodeId)
		return res, nil
	}
}

func (s *OrderService) GetById(ctx context.Context, req *m.ById) (*m.OrderStatisticRes, error) {
	order, err := s.storage.OrderS.GetById(req)
	if err != nil {
		return nil, err
	}

	return order, err
}
func (s *OrderService) GetAll(ctx context.Context, req *m.OrderGetListReq) (*m.PaginationResponse, *m.OrderList, error) {
	res := &m.PaginationResponse{}

	key := fmt.Sprintf("user_id:%s_course_id:%s_tariff_id:%s_promocode_id:%s_status:%s_payment_type:%s_type:%s_day:%s_from:%s_to:%s_page:%d_limit:%d",
		req.UserId, req.CourseId, req.TariffId, req.PromoCodeId, req.Status, req.PaymentType, req.Type, req.Day, req.From, req.To, req.Filter.Page, req.Filter.Limit)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res)
		}
	}

	if res.Total == 0 {
		res, _, err := s.storage.OrderS.GetAll(req)
		if err != nil {
			return nil, nil, err
		}

		if s.rdb != nil {
			if b, err := json.Marshal(res); err == nil {
				_ = s.rdb.Set(key, b, 60*time.Second).Err()
			}
		}

		return res, nil, nil
	}

	return res, nil, nil
}
func (s *OrderService) GetPayURL(ctx context.Context, req *m.ById) (*m.PayURLRes, error) {
	return s.storage.OrderS.GetPayURL(req)
}

func (s *OrderService) ApplyPrepare(ctx context.Context, req *m.ApplyPrepareParams) (*m.PrepareResult, error) {
	res, err := s.storage.OrderS.ApplyPrepare(req)
	fmt.Println("ApplyPrepare result:", res, "error:", err)
	fmt.Println("Request details:", req)
	fmt.Println("Response details:", res)
	return res, err
}
func (s *OrderService) ApplyComplete(ctx context.Context, req *m.ApplyCompleteParams) (*m.CompleteResult, error) {

	res, err := s.storage.OrderS.ApplyComplete(req)

	fmt.Println("ApplyComplete result:", res, "error:", err)
	fmt.Println("Request details:", req)
	fmt.Println("Response details:", res)

	return res, err
}
