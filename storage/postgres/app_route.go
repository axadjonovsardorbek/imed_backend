package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	m "imed/pkg/models"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type AppRouteRepo struct {
	db *sql.DB
}

func NewAppRouteRepo(db *sql.DB) *AppRouteRepo {
	return &AppRouteRepo{
		db: db,
	}
}

func (r *AppRouteRepo) Create(req *m.AppRouteCreateReq) (*m.ById, error) {
	id := uuid.NewString()

	appVersionBytes, err := json.Marshal(req.AppVersion)
	if err != nil {
		return nil, fmt.Errorf("marshal app_version: %w", err)
	}

	appLinksBytes, err := json.Marshal(req.AppLinks)
	if err != nil {
		return nil, fmt.Errorf("marshal app_links: %w", err)
	}
	query := `
	INSERT INTO app_routes (
		id,
		call_center,
		support_url,
		app_version,
		app_links,
		payment_min_version,
		feedback_url,
		buy_course
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.Exec(query,
		id,
		req.CallCenter,
		req.SupportUrl,
		appVersionBytes,
		appLinksBytes,
		req.PaymentMinVersion,
		req.FeedbackUrl,
		req.BuyCourse,
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating app route: %w", err)
	}

	return &m.ById{Id: id}, nil
}
func (r *AppRouteRepo) GetById(req *m.ById) (*m.AppRouteRes, error) {
	query := `
	SELECT 
		id,
		call_center,
		support_url,
		app_version,
		app_links,
		payment_min_version,
		buy_course,
		feedback_url,
		created_at,
		updated_at
	FROM
		app_routes
	WHERE
		deleted_at = 0
	AND
		id = $1
	`

	row := r.db.QueryRow(query, req.Id)

	app_route := m.AppRouteRes{}

	var rawAppVersion []byte
	var rawAppLinks []byte

	err := row.Scan(
		&app_route.Id,
		&app_route.CallCenter,
		&app_route.SupportUrl,
		&rawAppVersion,
		&rawAppLinks,
		&app_route.PaymentMinVersion,
		&app_route.BuyCourse,
		&app_route.FeedbackUrl,
		&app_route.CreatedAt,
		&app_route.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("app route not found with id: %s", req.Id)
	}

	if err != nil {
		return nil, fmt.Errorf("error while getting app route by id: %w", err)
	}

	if err := json.Unmarshal(rawAppVersion, &app_route.AppVersion); err != nil {
		return nil, fmt.Errorf("unmarshal app_version: %w", err)
	}

	if err := json.Unmarshal(rawAppLinks, &app_route.AppLinks); err != nil {
		return nil, fmt.Errorf("unmarshal app_links: %w", err)
	}

	return &app_route, nil
}
func (r *AppRouteRepo) GetAll(req *m.AppRouteGetListReq) (*m.PaginationResponse, *m.AppRouteList, error) {
	query := `
	SELECT 
		COUNT(id) OVER() AS total_count,
		id,
		call_center,
		support_url,
		app_version,
		app_links,
		payment_min_version,
		buy_course,
		feedback_url,
		created_at,
		updated_at
	FROM
		app_routes
	WHERE
		deleted_at = 0
	`

	var args []interface{}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting app route: %w", err)
	}
	defer rows.Close()

	app_routes := m.AppRouteList{}

	for rows.Next() {

		app_route := m.AppRouteRes{}

		var totalCount int

		var rawAppVersion []byte
		var rawAppLinks []byte

		err := rows.Scan(
			&totalCount,
			&app_route.Id,
			&app_route.CallCenter,
			&app_route.SupportUrl,
			&rawAppVersion,
			&rawAppLinks,
			&app_route.PaymentMinVersion,
			&app_route.BuyCourse,
			&app_route.FeedbackUrl,
			&app_route.CreatedAt,
			&app_route.UpdatedAt,
		)

		if err != nil {
			return nil, nil, fmt.Errorf("error while scanning app route row: %w", err)
		}

		if err := json.Unmarshal(rawAppVersion, &app_route.AppVersion); err != nil {
			return nil, nil, fmt.Errorf("unmarshal app_version: %w", err)
		}

		if err := json.Unmarshal(rawAppLinks, &app_route.AppLinks); err != nil {
			return nil, nil, fmt.Errorf("unmarshal app_links: %w", err)
		}

		app_routes.AppRoutes = append(app_routes.AppRoutes, &app_route)
		app_routes.Count = totalCount
	}

	response := m.NewPaginationResponse(m.Pagination{Limit: app_routes.Count}, app_routes.Count, app_routes.AppRoutes)

	return &response, &app_routes, nil
}
func (r *AppRouteRepo) Update(req *m.AppRouteUpdateReq) (*m.Void, error) {
	appVersionBytes, err := json.Marshal(req.AppVersion)
	if err != nil {
		return nil, fmt.Errorf("marshal app_version: %w", err)
	}

	appLinksBytes, err := json.Marshal(req.AppLinks)
	if err != nil {
		return nil, fmt.Errorf("marshal app_links: %w", err)
	}
	query := `
	UPDATE app_routes
	SET
	`

	var conditions []string
	var args []interface{}

	conditions = append(conditions, " call_center = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.CallCenter)

	conditions = append(conditions, " support_url = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.SupportUrl)

	conditions = append(conditions, " app_version = $"+strconv.Itoa(len(args)+1))
	args = append(args, appVersionBytes)

	conditions = append(conditions, " app_links = $"+strconv.Itoa(len(args)+1))
	args = append(args, appLinksBytes)

	conditions = append(conditions, " payment_min_version = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.PaymentMinVersion)

	conditions = append(conditions, " buy_course = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.BuyCourse)

	conditions = append(conditions, " feedback_url = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.FeedbackUrl)

	if len(conditions) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	conditions = append(conditions, " updated_at = now()")
	query += strings.Join(conditions, ", ")
	query += " WHERE id = $" + strconv.Itoa(len(args)+1) + " AND deleted_at = 0"

	args = append(args, req.Id)

	res, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error while updating app route: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("app route with id %s couldnt update", req.Id)
	}
	return &m.Void{}, nil
}
func (r *AppRouteRepo) Delete(req *m.ById) (*m.Void, error) {
	query := `
	UPDATE app_routes
	SET
		deleted_at = EXTRACT(EPOCH FROM NOW())
	WHERE
		id = $1
	AND
		deleted_at = 0
	`

	res, err := r.db.Exec(query, req.Id)
	if err != nil {
		return nil, fmt.Errorf("error while deleting app route: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("app route with id %s couldnt delete", req.Id)
	}

	return &m.Void{}, nil
}
