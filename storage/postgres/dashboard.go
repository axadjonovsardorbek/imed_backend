package postgres

import (
	"database/sql"
	"fmt"
	m "imed/pkg/models"
)

type DashboardRepo struct {
	db *sql.DB
}

func NewDashboardRepo(db *sql.DB) *DashboardRepo {
	return &DashboardRepo{
		db: db,
	}
}

func (r *DashboardRepo) GetDashboard(req *m.GetDashboardReq) (*m.DashboardRes, error) {
	query := `
	WITH bounds AS (
	  SELECT
		CASE
		  WHEN COALESCE($1,'') = '' THEN NULL
		  WHEN $1 = 'week'  THEN date_trunc('week',  now())
		  WHEN $1 = 'month' THEN date_trunc('month', now())
		  WHEN $1 = 'year'  THEN date_trunc('year',  now())
		  WHEN $1 = 'day' THEN date_trunc('day', COALESCE(NULLIF($2,'')::timestamptz, now()))
		  WHEN $1 = 'range' THEN COALESCE(NULLIF($3,'')::timestamptz, NULL)
		  ELSE NULL
		END AS from_ts,
		CASE
		  WHEN COALESCE($1,'') = '' THEN NULL
		  WHEN $1 IN ('week','month','year') THEN now()
		  WHEN $1 = 'day' THEN date_trunc('day', COALESCE(NULLIF($2,'')::timestamptz, now())) + interval '1 day'
		  WHEN $1 = 'range' THEN (COALESCE(NULLIF($4,'')::timestamptz, NULL) + interval '1 day')
		  ELSE NULL
		END AS to_ts
	)
	SELECT
	  (SELECT COUNT(*) FROM tariffs t, bounds b
		WHERE t.deleted_at=0
		  AND (b.from_ts IS NULL OR t.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR t.created_at <  b.to_ts)
	  ) AS tariffs,
	
	  (SELECT COUNT(*) FROM subjects s, bounds b
		WHERE s.deleted_at=0
		  AND (b.from_ts IS NULL OR s.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR s.created_at <  b.to_ts)
	  ) AS subjects,
	
	  (SELECT COUNT(*) FROM courses c, bounds b
		WHERE c.deleted_at=0
		  AND (b.from_ts IS NULL OR c.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR c.created_at <  b.to_ts)
	  ) AS courses,
	
	  (SELECT COUNT(*) FROM modules m, bounds b
		WHERE m.deleted_at=0
		  AND (b.from_ts IS NULL OR m.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR m.created_at <  b.to_ts)
	  ) AS modules,
	
	  (SELECT COUNT(*) FROM lessons l, bounds b
		WHERE l.deleted_at=0
		  AND (b.from_ts IS NULL OR l.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR l.created_at <  b.to_ts)
	  ) AS lessons,
	
	  (SELECT COUNT(*) FROM sources s, bounds b
		WHERE s.deleted_at=0 AND s.type='test'
		  AND (b.from_ts IS NULL OR s.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR s.created_at <  b.to_ts)
	  ) AS tests,
	
	  (SELECT COUNT(*) FROM sources s, bounds b
		WHERE s.deleted_at=0 AND s.type='document'
		  AND (b.from_ts IS NULL OR s.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR s.created_at <  b.to_ts)
	  ) AS documents,
	
	  (SELECT COUNT(*) FROM sources s, bounds b
		WHERE s.deleted_at=0 AND s.type='video'
		  AND (b.from_ts IS NULL OR s.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR s.created_at <  b.to_ts)
	  ) AS videos,
	
	  (SELECT COUNT(*) FROM users u, bounds b
		WHERE u.deleted_at=0
		  AND (b.from_ts IS NULL OR u.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR u.created_at <  b.to_ts)
	  ) AS users,
	
	  (SELECT COUNT(*) FROM teachers t, bounds b
		WHERE t.deleted_at=0
		  AND (b.from_ts IS NULL OR t.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR t.created_at <  b.to_ts)
	  ) AS teachers,
	
	  (SELECT COUNT(DISTINCT ua.user_id)
		FROM user_activities ua
		JOIN users u ON u.id = ua.user_id
		CROSS JOIN bounds b
		WHERE u.deleted_at = 0
		  AND ua.activity <> 0
		  AND (b.from_ts IS NULL OR ua.day >= b.from_ts::date)
		  AND (b.to_ts   IS NULL OR ua.day <  b.to_ts::date)
	  ) AS active_users,
	  
	  -- User courses
	  (SELECT COUNT(*) FROM user_courses uc, bounds b
		WHERE uc.deleted_at = 0
		  AND (b.from_ts IS NULL OR uc.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR uc.created_at < b.to_ts)
	  ) AS user_courses,
	
	  -- Active user courses
	  (SELECT COUNT(*) FROM user_courses uc, bounds b
		WHERE uc.deleted_at = 0
		  AND uc.started_at <= now() 
		  AND uc.ended_at > now()
		  AND (b.from_ts IS NULL OR uc.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR uc.created_at < b.to_ts)
	  ) AS active_user_courses,
	  
	  -- Inactive user courses
	  (SELECT COUNT(*) FROM user_courses uc, bounds b
		WHERE uc.deleted_at = 0
		  AND uc.ended_at <= now()
		  AND (b.from_ts IS NULL OR uc.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR uc.created_at < b.to_ts)
	  ) AS inactive_user_courses,
	
	  (SELECT COUNT(*) FROM user_courses uc, bounds b
		WHERE uc.deleted_at <> 0
		  AND (b.from_ts IS NULL OR uc.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR uc.created_at < b.to_ts)
	  ) AS deleted_user_courses,
	
	  -- Orders
	  (SELECT COUNT(*) FROM orders o, bounds b
		WHERE o.status='PAID'
		  AND (b.from_ts IS NULL OR o.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR o.created_at < b.to_ts)
	  ) AS paid_orders,
	
	  (SELECT COUNT(*) FROM orders o, bounds b
		WHERE o.payment_type='admin' AND o.status='PAID'
		  AND (b.from_ts IS NULL OR o.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR o.created_at < b.to_ts)
	  ) AS admin_orders,
	
	  (SELECT COUNT(*) FROM orders o, bounds b
		WHERE o.payment_type='click' AND o.status='PAID'
		  AND (b.from_ts IS NULL OR o.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR o.created_at < b.to_ts)
	  ) AS click_orders,
	
	  (SELECT COUNT(*) FROM orders o, bounds b
		WHERE o.status='CANCELLED'
		  AND (b.from_ts IS NULL OR o.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR o.created_at < b.to_ts)
	  ) AS cancelled_orders,
	
	  -- Promo codes
	  (SELECT COUNT(*) FROM promocodes p, bounds b
		WHERE p.deleted_at = 0
		  AND (b.from_ts IS NULL OR p.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR p.created_at < b.to_ts)
	  ) AS promocodes,
	
	  (SELECT COUNT(*) FROM promocodes p, bounds b
		WHERE p.deleted_at = 0 AND p.is_active = TRUE
		  AND (b.from_ts IS NULL OR p.created_at >= b.from_ts)
		  AND (b.to_ts   IS NULL OR p.created_at < b.to_ts)
	  ) AS active_promocodes;
	`

	row := r.db.QueryRow(query, req.Type, req.Day, req.From, req.To)

	var dashboard m.DashboardRes
	err := row.Scan(
		&dashboard.Tariffs,
		&dashboard.Subjects,
		&dashboard.Courses,
		&dashboard.Modules,
		&dashboard.Lessons,
		&dashboard.Tests,
		&dashboard.Documents,
		&dashboard.Videos,
		&dashboard.Users,
		&dashboard.Teachers,
		&dashboard.ActiveUsers,
		&dashboard.UserCourses,
		&dashboard.ActiveUserCourses,
		&dashboard.InactiveUserCourses,
		&dashboard.DeletedUserCourses,
		&dashboard.PaidOrders,
		&dashboard.AdminOrders,
		&dashboard.ClickOrders,
		&dashboard.CancelledOrders,
		&dashboard.PromoCodes,
		&dashboard.ActivePromoCodes,
	)
	if err != nil {
		return nil, fmt.Errorf("error while getting dashboard: %w", err)
	}

	return &dashboard, nil
}
