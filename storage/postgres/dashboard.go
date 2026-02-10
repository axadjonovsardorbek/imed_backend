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
	  ) AS teachers
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
	)
	if err != nil {
		return nil, fmt.Errorf("error while getting dashboard: %w", err)
	}

	return &dashboard, nil
}
