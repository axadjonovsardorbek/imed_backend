package service

import (
	"context"
	"encoding/json"
	"fmt"
	"imed/internal/helper"
	m "imed/pkg/models"
	st "imed/storage/postgres"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type OrderService struct {
	storage st.Storage
	rdb     *redis.Client
	push    *PushService
}

func NewOrderService(storage *st.Storage, rdb *redis.Client, push *PushService) *OrderService {
	return &OrderService{storage: *storage, rdb: rdb, push: push}
}

func (s *OrderService) getVersion() string {
	version := "1"
	if s.rdb != nil {
		if v, err := s.rdb.Get(context.Background(), "order:version").Result(); err == nil && v != "" {
			version = v
		}
	}
	return version
}

func (s *OrderService) Create(ctx context.Context, req *m.OrderCreateReq) (*m.OrderCreateRes, error) {

	var (
		res *m.OrderCreateRes
		err error
	)

	if req.PromoCodeId == "" || req.PromoCodeId == "string" {
		res, err = s.storage.OrderS.Create(req)
	} else {
		res, err = s.storage.OrderS.CreateWithPromoCode(req)
	}

	if err != nil {
		return nil, err
	}

	// 🔥 cache invalidation
	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "order:version").Err()
	}

	return res, nil
}

func (s *OrderService) GetById(ctx context.Context, req *m.ById) (*m.OrderStatisticRes, error) {
	order, err := s.storage.OrderS.GetById(req)
	if err != nil {
		return nil, err
	}

	return order, err
}
func (s *OrderService) GetAll(ctx context.Context, req *m.OrderGetListReq) (*m.PaginationResponse, *m.OrderList, error) {

	version := s.getVersion()

	key := fmt.Sprintf(
		"order:list:v%s_user:%s_course:%s_tariff:%s_promo:%s_status:%s_pay:%s_type:%s_day:%s_from:%s_to:%s_page:%d_limit:%d",
		version,
		req.UserId,
		req.CourseId,
		req.TariffId,
		req.PromoCodeId,
		req.Status,
		req.PaymentType,
		req.Type,
		req.Day,
		req.From,
		req.To,
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

	// 🔥 GET FROM DB
	res, list, err := s.storage.OrderS.GetAll(req)
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

func (s *OrderService) GetPayURL(ctx context.Context, req *m.ById) (*m.PayURLRes, error) {
	return s.storage.OrderS.GetPayURL(req)
}

func (s *OrderService) ApplyPrepare(ctx context.Context, req *m.ApplyPrepareParams) (*m.PrepareResult, error) {
	res, err := s.storage.OrderS.ApplyPrepare(req)
	// req to'liq chop etilmaydi: MetaPatchJSON ichida Click imzosi bor
	// (baytlar ko'rinishida bo'lsa ham o'qib olinadi).
	slog.Info("click apply prepare", "order_id", req.OrderId, "result", res, "err", err)
	return res, err
}
func (s *OrderService) ApplyComplete(ctx context.Context, req *m.ApplyCompleteParams) (*m.CompleteResult, error) {

	res, err := s.storage.OrderS.ApplyComplete(req)

	// ApplyPrepare'dagi kabi: req to'liq chop etilmaydi.
	slog.Info("click apply complete", "order_id", req.OrderId, "result", res, "err", err)

	orderID := req.OrderId
	code := res.Error

	go func(orderID string, code int) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("push goroutine panicked", "panic", r)
			}
		}()

		pushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		order, err2 := s.storage.OrderS.GetById(&m.ById{Id: orderID})
		if err2 != nil || order == nil || *order.UserId == "" {
			slog.Error("push skipped, order fetch failed", "err", err2)
			return
		}

		var push_event m.PushEvent
		if code == 0 {
			push_event = m.PushEvent{
				Type:       m.EventPaymentSuccess,
				OrderID:    order.Id,
				CourseID:   *order.CourseId,
				CourseName: *order.CourseName,
			}
		} else if code != 0 && code != -4 {
			push_event = m.PushEvent{
				Type:       m.EventPaymentFailed,
				OrderID:    order.Id,
				CourseID:   *order.CourseId,
				CourseName: *order.CourseName,
			}
		}

		push_target := m.PushTarget{Type: m.TargetUser, UserID: *order.UserId}

		err = s.push.SendPush(ctx,
			push_target,
			push_event,
		)

		if err != nil {
			slog.Error("send push failed", "err", err)
		}

		msg := m.BuildUzPushMessage(push_event)
		title, body := helper.StringToJSON(msg.Title, msg.Body)

		ntf, err := s.storage.NotificationS.Create(&m.NotificationCreateReq{
			Title:    title,
			Message:  body,
			Type:     "selected",
			CourseId: *order.CourseId,
			UserId:   *order.UserId,
		})

		if err != nil {
			slog.Error("create notification failed", "err", err)
		}

		slog.Info("order push sent",
			"order_id", orderID,
			"payment_code", code,
			"event", push_event,
			"notification_id", ntf.Id,
		)

		if err := s.push.SendPush(pushCtx,
			push_target,
			push_event,
		); err != nil {
			slog.Error("push send failed", "err", err)
		}

	}(orderID, code)

	return res, err
}
