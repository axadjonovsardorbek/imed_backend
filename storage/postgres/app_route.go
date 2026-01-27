package postgres

import (
	"database/sql"
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
	query := `
	INSERT INTO app_routes (
		id,
		key,
		value,
		description,
		order_num
	) VALUES ($1, $2, $3::jsonb, $4, $5)
	`

	_, err := r.db.Exec(query,
		id,
		req.Key,
		[]byte(req.Value),
		req.Description,
		req.OrderNum,
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
		key,
		value,
		description,
		order_num,
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
	err := row.Scan(
		&app_route.Id,
		&app_route.Key,
		&app_route.Value,
		&app_route.Description,
		&app_route.OrderNum,
		&app_route.CreatedAt,
		&app_route.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("app route not found with id: %s", req.Id)
	}

	if err != nil {
		return nil, fmt.Errorf("error while getting app route by id: %w", err)
	}

	return &app_route, nil
}
func (r *AppRouteRepo) GetAll(req *m.AppRouteGetListReq) (*m.PaginationResponse, *m.AppRouteList, error) {
	query := `
	SELECT 
		COUNT(id) OVER() AS total_count,
		id,
		key,
		value,
		description,
		order_num,
		created_at,
		updated_at
	FROM
		app_routes
	WHERE
		deleted_at = 0
	`

	var args []interface{}
	if req.Key != "" && req.Key != "string" {
		query += fmt.Sprintf(" AND key ILIKE $%d", len(args)+1)
		args = append(args, "%"+req.Key+"%")
	}

	query += " ORDER BY order_num ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting app route: %w", err)
	}
	defer rows.Close()

	app_routes := m.AppRouteList{}

	for rows.Next() {

		app_route := m.AppRouteRes{}

		var totalCount int

		err := rows.Scan(
			&totalCount,
			&app_route.Id,
			&app_route.Key,
			&app_route.Value,
			&app_route.Description,
			&app_route.OrderNum,
			&app_route.CreatedAt,
			&app_route.UpdatedAt,
		)

		if err != nil {
			return nil, nil, fmt.Errorf("error while scanning app route row: %w", err)
		}

		app_routes.AppRoutes = append(app_routes.AppRoutes, &app_route)
		app_routes.Count = totalCount
	}

	response := m.NewPaginationResponse(m.Pagination{Limit: app_routes.Count}, app_routes.Count, app_routes.AppRoutes)

	return &response, &app_routes, nil
}
func (r *AppRouteRepo) Update(req *m.AppRouteUpdateReq) (*m.Void, error) {
	query := `
	UPDATE app_routes
	SET
	`

	var conditions []string
	var args []interface{}

	if req.Key != "" && req.Key != "string" {
		conditions = append(conditions, " key = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.Key)
	}

	conditions = append(conditions, " description = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.Description)

	if req.Value != nil && len(req.Value) > 0 {
		args = append(args, []byte(req.Value))
		conditions = append(conditions, "value = $"+strconv.Itoa(len(args))+"::jsonb")
	}

	conditions = append(conditions, " order_num = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.OrderNum)

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
