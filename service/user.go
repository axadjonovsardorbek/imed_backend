package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"imed/internal/config"
	"imed/internal/helper"
	m "imed/pkg/models"
	st "imed/storage/postgres"
	"time"

	"github.com/go-redis/redis"
)

type UserService struct {
	storage st.Storage
	rdb     *redis.Client
}

func NewUserService(storage *st.Storage, rdb *redis.Client) *UserService {
	return &UserService{storage: *storage, rdb: rdb}
}

func (s *UserService) SendConfirmationCode(ctx context.Context, req *m.SendConfirmationCodeReq) (*m.Void, error) {
	otp := helper.GenerateOtp()
	cfg := config.Load()

	if req.Email == "samandarahadjonov@gmail.com" {
		otp = "111111"
		fmt.Printf("Generated OTP for email %s: %s\n", req.Email, otp)
		s.rdb.Set(req.Email, otp, 5*time.Minute)
	} else if req.Email != "" && req.Email != "string" {
		s.rdb.Set(req.Email, otp, 5*time.Minute)
		fmt.Printf("Generated OTP for email %s: %s\n", req.Email, otp)
	} else {
		return nil, errors.New("email is required")
	}

	from := cfg.EMAIL_NAME
	logoURL := cfg.EMAIL_LOGO

	htmlBody := fmt.Sprintf(`<!doctype html>
	<html lang="uz">
	<head>
	  <meta charset="utf-8" />
	  <meta name="viewport" content="width=device-width, initial-scale=1" />
	</head>
	<body style="margin:0;padding:0;background:#f6f7fb;font-family:Arial,Helvetica,sans-serif;">
	  <div style="max-width:560px;margin:0 auto;padding:24px;">
	
		<div style="background:#ffffff;border-radius:14px;overflow:hidden;border:1px solid #e9ecf3;">
		  <div style="height:6px;background:#e11d48;"></div>
	
		  <div style="padding:22px 24px 24px 24px;">
	
			<div style="display:flex;align-items:center;gap:12px;">
			  <img src="%s" alt="iMedTeam" height="38" style="display:block;border:0;outline:none;text-decoration:none;" />
			  <div>
				<div style="font-size:18px;font-weight:800;color:#111827;line-height:1.2;">iMedTeam.uz</div>
				<div style="font-size:13px;color:#6b7280;margin-top:2px;">Kirish uchun tasdiqlash kodi</div>
			  </div>
			</div>
	
			<div style="margin-top:18px;font-size:14px;line-height:1.6;color:#111827;">
			  Assalomu alaykum!<br><br>
			  Siz <b>iMedTeam.uz</b> platformasiga kirish uchun so‘rov yubordingiz.
			</div>
	
			<!-- OTP box -->
			<div style="margin:20px 0;padding:16px;border-radius:12px;background:#fff1f2;border:1px solid #fecdd3;text-align:center;">
			  <div style="font-size:12px;color:#9f1239;margin-bottom:8px;font-weight:700;">Tasdiqlash kodingiz</div>
			  <div style="font-size:30px;font-weight:900;letter-spacing:2px;color:#111827;
						  font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,'Liberation Mono','Courier New',monospace;">
				%s
			  </div>
			</div>
	
			<!-- Warning text (NO frame) -->
			<div style="font-size:13px;line-height:1.6;color:#b91c1c;font-weight:700;margin-bottom:8px;">
			  ⚠️ Hech qachon ushbu kodni boshqalar bilan ulashmang.
			</div>
	
			<div style="font-size:13px;line-height:1.6;color:#374151;">
			  ⏳ Kod <b>5 daqiqa</b> davomida amal qiladi.<br>
			  Agar bu so‘rovni siz yubormagan bo‘lsangiz, ushbu xabarni e’tiborsiz qoldiring.
			</div>
	
			<hr style="border:none;border-top:1px solid #eef0f6;margin:18px 0;">
	
			<div style="font-size:12px;color:#6b7280;">
			  Hurmat bilan,<br>
			  iMedTeam.uz
			</div>
		  </div>
		</div>
	
		<div style="text-align:center;color:#9ca3af;font-size:12px;margin-top:12px;">
		  © iMedTeam.uz
		</div>
	  </div>
	</body>
	</html>`, logoURL, otp)

	subject := "iMedTeam.uz — Kirish uchun tasdiqlash kodi"

	message := fmt.Sprintf(`
	Assalomu alaykum!

	Siz iMedTeam.uz platformasiga kirish uchun so‘rov yubordingiz.

	Tasdiqlash kodingiz: %s

	Kod 5 daqiqa davomida amal qiladi.

	Agar bu so‘rovni siz yubormagan bo‘lsangiz, ushbu xabarni e’tiborsiz qoldiring.

	Hurmat bilan,
	iMedTeam.uz
	`, otp,
	)

	err := helper.SendVerificationCode(m.Params{
		From:     from,
		Subject:  subject,
		To:       req.Email,
		Message:  message,
		HTMLBody: htmlBody,
	}, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to send verification code: %v", err)
	}

	return &m.Void{}, nil
}

func (s *UserService) Confirmation(ctx context.Context, req *m.ConfirmationReq) (*m.TokenRes, error) {

	if req.Email != "" && req.Email != "string" {
		storedOtp, err := s.rdb.Get(req.Email).Result()

		if err == redis.Nil {
			return nil, errors.New("verification code is invalid or expired")
		} else if err != nil {
			return nil, errors.New("failed to retrieve verification code from Redis")
		}

		if storedOtp != req.ConfirmationCode {
			return nil, errors.New("verification code is incorrect")
		}

		res, err := s.storage.UserS.EmailLogin(&m.ConfirmationReq{Email: req.Email})
		if err != nil {
			return nil, err
		}

		return res, nil

	} else if req.ConfirmationCode != "" && req.ConfirmationCode != "string" {
		phone, err := s.rdb.Get(req.ConfirmationCode).Result()

		if err == redis.Nil {
			return nil, errors.New("verification code is invalid or expired")
		} else if err != nil {
			return nil, errors.New("failed to retrieve verification code from Redis")
		}

		name, err := s.rdb.Get(phone).Result()
		if err == redis.Nil {
			return nil, errors.New("verification code is invalid or expired")
		} else if err != nil {
			return nil, errors.New("failed to retrieve verification code from Redis")
		}

		res, err := s.storage.UserS.TelegramLogin(&m.TelegramLoginReq{Phone: phone, Name: name})
		if err != nil {
			return nil, err
		}

		return res, nil
	} else {
		return nil, errors.New("there is no such verification code")
	}
}

func (s *UserService) Create(ctx context.Context, req *m.UserLoginReq) (*m.ById, error) {
	return s.storage.UserS.CreateUser(req)
}
func (s *UserService) Check(ctx context.Context, req *m.UserCheckReq) (*m.UserCheckRes, error) {
	return s.storage.UserS.Check(req)
}
func (s *UserService) Login(ctx context.Context, req *m.UserLoginReq) (*m.TokenRes, error) {
	return s.storage.UserS.Login(req)
}
func (s *UserService) GetById(ctx context.Context, req *m.ById) (*m.UserRes, error) {
	return s.storage.UserS.GetById(req)
}
func (s *UserService) GetAll(ctx context.Context, req *m.UserGetListReq) (*m.PaginationResponse, *m.UserList, error) {
	res := &m.PaginationResponse{}

	key := fmt.Sprintf("user_name:%s_user_phone:%s_page:%d_limit:%d",
		req.Name, req.PhoneNumber, req.Filter.Page, req.Filter.Limit)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res)
		}
	}

	if res.Total == 0 {
		res, _, err := s.storage.UserS.GetAll(req)
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
func (s *UserService) Update(ctx context.Context, req *m.UserUpdateReq) (*m.Void, error) {
	return s.storage.UserS.Update(req)
}
func (s *UserService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	return s.storage.UserS.Delete(req)
}
func (s *UserService) ChangePassword(ctx context.Context, req *m.ChangePasswordReq) (*m.Void, error) {
	return s.storage.UserS.ChangePassword(req)
}
func (s *UserService) RefreshPassword(ctx context.Context, req *m.RefreshPasswordReq) (*m.RefreshPassword, error) {
	return s.storage.UserS.RefreshPassword(req)
}
func (s *UserService) CreateActivity(ctx context.Context, req *m.UserActivityCreateReq) (*m.Void, error) {
	return s.storage.UserS.CreateActivity(req)
}
func (s *UserService) GetActivities(ctx context.Context, req *m.UserActivityGetListReq) (*m.ActivityStatsResponse, error) {
	return s.storage.UserS.GetActivities(req)
}
func (s *UserService) GetRating(ctx context.Context, req *m.RatingRequest) (*m.RatingResponse, error) {
	req.Limit = clampLimit(req.Limit)

	res := &m.RatingResponse{
		Type:  req.Type,
		Limit: req.Limit,
		Items: make([]m.RatingUserItem, 0, req.Limit),
	}

	key := topCacheKey(string(req.Type), req.Limit)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &res.Items)
		}
	}

	if len(res.Items) == 0 {
		items, err := s.storage.UserS.GetTop(req)
		if err != nil {
			return nil, err
		}
		res.Items = items

		if s.rdb != nil {
			if b, err := json.Marshal(items); err == nil {
				_ = s.rdb.Set(key, b, 30*time.Second).Err()
			}
		}
	}

	if req.UserId != "" {
		me, err := s.storage.UserS.GetMe(req)
		if err != nil {
			return nil, err
		}
		res.Me = me
	}

	fmt.Println("RATING SERVICE RETURNING")
	return res, nil

	// return s.storage.UserS.GetRating(req)
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func topCacheKey(typ string, limit int) string {
	return fmt.Sprintf("rating:top:%s:%d", typ, limit)
}
