package service

import (
	"context"
	"encoding/json"
	"fmt"
	"imed/internal/helper"
	m "imed/pkg/models"
	st "imed/storage/postgres"
	"time"

	"github.com/redis/go-redis/v9"
)

type ModuleService struct {
	storage st.Storage
	rdb     *redis.Client
	push    *PushService
}

func NewModuleService(storage *st.Storage, rdb *redis.Client, push *PushService) *ModuleService {
	return &ModuleService{storage: *storage, rdb: rdb, push: push}
}

func (s *ModuleService) getVersion() string {
	version := "1"
	if s.rdb != nil {
		if v, err := s.rdb.Get(context.Background(), "module:version").Result(); err == nil && v != "" {
			version = v
		}
	}
	return version
}

func (s *ModuleService) Create(ctx context.Context, req *m.ModuleCreateReq) (*m.ById, error) {
	res, err := s.storage.ModuleS.Create(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "module:version").Err()
	}

	info, err := s.storage.CourseS.GetById(&m.ById{Id: req.CourseId})
	if err == nil {
		push_target := m.PushTarget{Type: m.TargetAllUsers}

		nameUz, _ := helper.PickFromJSONStr(req.Name, "uz")

		push_event := m.PushEvent{
			Type:       m.EventModuleAdded,
			CourseID:   info.Id,
			CourseName: info.Name["uz"],
			ModuleName: nameUz,
			ModuleID:   res.Id,
		}

		_ = s.push.SendPush(ctx, push_target, push_event)

		msg := m.BuildUzPushMessage(push_event)
		title, body := helper.StringToJSON(msg.Title, msg.Body)

		_, _ = s.storage.NotificationS.Create(&m.NotificationCreateReq{
			Title:   title,
			Message: body,
			Type:    "all",
		})
	}

	return res, nil
}

func (s *ModuleService) GetById(ctx context.Context, req *m.ById) (*m.ModuleRes, error) {
	return s.storage.ModuleS.GetById(req)
}

func (s *ModuleService) GetAll(ctx context.Context, req *m.ModuleGetListReq) (*m.PaginationResponse, *m.ModuleList, error) {

	version := s.getVersion()

	key := fmt.Sprintf(
		"module:list:v%s_course:%s_name:%s_page:%d_limit:%d",
		version,
		req.CourseId,
		req.Name,
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

	res, list, err := s.storage.ModuleS.GetAll(req)
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

func (s *ModuleService) Update(ctx context.Context, req *m.ModuleUpdateReq) (*m.Void, error) {
	res, err := s.storage.ModuleS.Update(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "module:version").Err()
	}

	module, err := s.storage.ModuleS.GetById(&m.ById{Id: req.Id})
	if err == nil && module.CourseId != nil {
		info, err := s.storage.CourseS.GetById(&m.ById{Id: *module.CourseId})
		if err == nil {
			push_target := m.PushTarget{Type: m.TargetAllUsers}

			nameUz, _ := helper.PickFromJSONStr(req.Name, "uz")

			push_event := m.PushEvent{
				Type:       m.EventModuleUpdated,
				CourseID:   info.Id,
				CourseName: info.Name["uz"],
				ModuleName: nameUz,
				ModuleID:   module.Id,
			}

			_ = s.push.SendPush(ctx, push_target, push_event)

			msg := m.BuildUzPushMessage(push_event)
			title, body := helper.StringToJSON(msg.Title, msg.Body)

			_, _ = s.storage.NotificationS.Create(&m.NotificationCreateReq{
				Title:   title,
				Message: body,
				Type:    "all",
			})
		}
	}

	return res, nil
}

func (s *ModuleService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	res, err := s.storage.ModuleS.Delete(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "module:version").Err()
	}

	return res, nil
}
