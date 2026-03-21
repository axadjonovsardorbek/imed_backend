package service

import (
	"context"
	"encoding/json"
	"fmt"
	"imed/internal/helper"
	m "imed/pkg/models"
	st "imed/storage/postgres"
	"time"

	"github.com/go-redis/redis"
)

type OrderService struct {
	storage st.Storage
	rdb     *redis.Client
	push    *PushService
}

func NewOrderService(storage *st.Storage, rdb *redis.Client, push *PushService) *OrderService {
	return &OrderService{storage: *storage, rdb: rdb, push: push}
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

	orderID := req.OrderId
	code := res.Error

	go func(orderID string, code int) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("push goroutine panic:", r)
			}
		}()

		pushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		order, err2 := s.storage.OrderS.GetById(&m.ById{Id: orderID})
		if err2 != nil || order == nil || order.UserId == "" {
			fmt.Println("push skipped, order fetch error:", err2)
			return
		}

		var push_event m.PushEvent
		if code == 0 {
			push_event = m.PushEvent{
				Type:       m.EventPaymentSuccess,
				OrderID:    order.Id,
				CourseID:   order.CourseId,
				CourseName: order.CourseName,
			}
		} else if code != 0 && code != -4 {
			push_event = m.PushEvent{
				Type:       m.EventPaymentFailed,
				OrderID:    order.Id,
				CourseID:   order.CourseId,
				CourseName: order.CourseName,
			}
		}

		push_target := m.PushTarget{Type: m.TargetUser, UserID: order.UserId}

		err = s.push.SendPush(ctx,
			push_target,
			push_event,
		)

		if err != nil {
			fmt.Println("Failed to send push:", err)
		}

		msg := m.BuildUzPushMessage(push_event)
		title, body := helper.StringToJSON(msg.Title, msg.Body)

		ntf, err := s.storage.NotificationS.Create(&m.NotificationCreateReq{
			Title:    title,
			Message:  body,
			Type:     "selected",
			CourseId: order.CourseId,
			UserId:   order.UserId,
		})

		if err != nil {
			fmt.Println("Failed to create notification:", err)
		}

		fmt.Println("Push event sent:", push_event)
		fmt.Println("Push message:", msg)
		fmt.Println("Order ID:", orderID, "Payment code:", code)
		fmt.Println("Notification created with ID:", ntf.Id)

		if err := s.push.SendPush(pushCtx,
			push_target,
			push_event,
		); err != nil {
			fmt.Println("push send error:", err)
		}

	}(orderID, code)

	return res, err
}
