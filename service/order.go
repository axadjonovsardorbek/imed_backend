package service

import (
	"context"
	"fmt"
	"imed/internal/helper"
	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type OrderService struct {
	storage st.Storage
}

func NewOrderService(storage *st.Storage) *OrderService {
	return &OrderService{storage: *storage}
}

func (s *OrderService) Create(ctx context.Context, req *m.OrderCreateReq) (*m.OrderCreateRes, error) {
	return s.storage.OrderS.Create(req)
}
func (s *OrderService) GetById(ctx context.Context, req *m.ById) (*m.OrderRes, error) {
	return s.storage.OrderS.GetById(req)
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
	if err != nil {
		return nil, err
	}
	fmt.Println("ApplyComplete result:", res, "error:", err)
	fmt.Println("Request details:", req)
	fmt.Println("Response details:", res)

	if res.Error == 0 {
		order, _ := s.storage.OrderS.GetById(&m.ById{Id: req.OrderId})

		tariff, _ := s.storage.TariffS.GetById(&m.ById{Id: order.TariffId})

		start, end := helper.GetStartEndDate(*tariff.Duration)

		req_course := m.UserCourseCreateReq{
			UserId:    order.UserId,
			CourseId:  order.CourseId,
			TariffId:  order.TariffId,
			StartedAt: start,
			EndedAt:   end,
		}
		_, err := s.storage.CourseS.BuyCourse(&req_course)
		if err == nil {
			fmt.Println("Course successfully added to user after order completion:", order.UserId, order.CourseId)
			// return res, nil
		}

	}

	return res, nil
}
