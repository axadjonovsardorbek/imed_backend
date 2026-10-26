package service

import (
	"context"
	"encoding/json"
	"fmt"
	m "imed/pkg/models"
	st "imed/storage/postgres"
	"time"

	"log/slog"

	"github.com/redis/go-redis/v9"
)

type BannerService struct {
	storage st.Storage
	rdb     *redis.Client
	push    *PushService
}

func NewBannerService(storage *st.Storage, rdb *redis.Client, push *PushService) *BannerService {
	return &BannerService{storage: *storage, rdb: rdb, push: push}
}

func (s *BannerService) getVersion() string {
	version := "1"
	if s.rdb != nil {
		if v, err := s.rdb.Get(context.Background(), "banner:version").Result(); err == nil && v != "" {
			version = v
		}
	}
	return version
}

func (s *BannerService) Create(ctx context.Context, req *m.BannerCreateReq) (*m.ById, error) {
	res, err := s.storage.BannerS.Create(req)
	if err != nil {
		return nil, err
	}
	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "banner:version").Err()
	}

	bannerTitleJSON := `{"uz": "Yangi banner qo'shildi 🔥", "ru": "Добавлен новый баннер 🔥", "en": "New banner added 🔥"}`
	_ = s.push.SendPush(ctx, m.PushTarget{Type: m.TargetAllUsers}, m.PushEvent{
		Type:        m.EventBannerAdded,
		TitleJSON:   bannerTitleJSON,
		MessageJSON: req.Title,
	})

	return res, nil
}

func (s *BannerService) Update(ctx context.Context, req *m.BannerUpdateReq) (*m.Void, error) {
	res, err := s.storage.BannerS.Update(req)
	if err != nil {
		return nil, err
	}
	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "banner:version").Err()
	}

	bannerTitleJSON := `{"uz": "Banner yangilandi 🔄", "ru": "Баннер обновлен 🔄", "en": "Banner updated 🔄"}`
	_ = s.push.SendPush(ctx, m.PushTarget{Type: m.TargetAllUsers}, m.PushEvent{
		Type:        m.EventBannerAdded,
		TitleJSON:   bannerTitleJSON,
		MessageJSON: req.Title,
	})

	return res, nil
}

func (s *BannerService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	res, err := s.storage.BannerS.Delete(req)
	if err != nil {
		return nil, err
	}
	if s.rdb != nil {
		_ = s.rdb.Incr(context.Background(), "banner:version").Err()
	}
	return res, nil
}

func (s *BannerService) GetById(ctx context.Context, req *m.ById) (*m.BannerRes, error) {
	return s.storage.BannerS.GetById(req)
}

func (s *BannerService) GetByIdMobile(ctx context.Context, req *m.ByIdMobile) (*m.BannerMobileRes, error) {
	return s.storage.BannerS.GetByIdMobile(req)
}

func (s *BannerService) GetAll(ctx context.Context, req *m.BannerGetListReq) (*m.PaginationResponse, error) {
	version := s.getVersion()

	key := fmt.Sprintf("banner:list:v%s:title:%s_page:%d_limit:%d",
		version,
		req.Title,
		req.Filter.Page,
		req.Filter.Limit,
	)

	if s.rdb != nil {
		if b, err := s.rdb.Get(context.Background(), key).Bytes(); err == nil && len(b) > 0 {
			var cached m.PaginationResponse
			if err := json.Unmarshal(b, &cached); err == nil {
				slog.Info("Cache Hit", "service", "banner", "key", key)
				return &cached, nil
			}
		}
	}

	slog.Info("Cache Miss", "service", "banner", "key", key)

	res, _, err := s.storage.BannerS.GetAll(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		if b, err := json.Marshal(res); err == nil {
			_ = s.rdb.Set(context.Background(), key, b, 5*time.Minute).Err()
		}
	}

	return res, nil
}

func (s *BannerService) GetAllMobile(ctx context.Context, req *m.BannerGetListMobileReq) (*m.PaginationResponse, error) {
	version := s.getVersion()

	key := fmt.Sprintf("banner:mobile:list:v%s", version)

	if s.rdb != nil {
		if b, err := s.rdb.Get(context.Background(), key).Bytes(); err == nil && len(b) > 0 {
			var cached m.PaginationResponse
			if err := json.Unmarshal(b, &cached); err == nil {
				slog.Info("Cache Hit (Mobile)", "service", "banner", "key", key)
				return &cached, nil
			}
		}
	}

	slog.Info("Cache Miss (Mobile)", "service", "banner", "key", key)

	res, _, err := s.storage.BannerS.GetAllMobile(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		if b, err := json.Marshal(res); err == nil {
			_ = s.rdb.Set(context.Background(), key, b, 5*time.Minute).Err()
		}
	}

	return res, nil
}
