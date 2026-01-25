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

type ModuleRepo struct {
	db *sql.DB
}

func NewModuleRepo(db *sql.DB) *ModuleRepo {
	return &ModuleRepo{
		db: db,
	}
}

func (r *ModuleRepo) Create(req *m.ModuleCreateReq) (*m.ById, error) {
	id := uuid.NewString()
	query := `
	INSERT INTO modules (
		id,
		name,
		course_id,
		order_num
	) VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(query,
		id,
		req.Name,
		req.CourseId,
		req.OrderNum,
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating module: %w", err)
	}

	return &m.ById{Id: id}, nil
}

func (r *ModuleRepo) GetById(req *m.ById) (*m.ModuleRes, error) {

	query := `
	SELECT 
		m.id,
		m.name,
		m.order_num,
		m.course_id,
		c.name,
		m.created_at,
		m.updated_at
	FROM
		modules m
	LEFT JOIN
		courses c ON m.course_id = c.id
	WHERE
		m.deleted_at = 0
	AND
		m.id = $1
	`

	row := r.db.QueryRow(query, req.Id)

	module := m.ModuleRes{}
	var nameJSON []byte
	var courseNameJSON []byte

	err := row.Scan(
		&module.Id,
		&nameJSON,
		&module.OrderNum,
		&module.CourseId,
		&courseNameJSON,
		&module.CreatedAt,
		&module.UpdatedAt,
	)

	json.Unmarshal(nameJSON, &module.Name)
	json.Unmarshal(courseNameJSON, &module.CourseName)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("module not found with id: %s", req.Id)
	}
	if err != nil {
		return nil, fmt.Errorf("error while getting module by id: %w", err)
	}
	return &module, nil
}

func (r *ModuleRepo) GetAll(req *m.ModuleGetListReq) (*m.PaginationResponse, *m.ModuleList, error) {
	query := `
	SELECT 
		COUNT(m.id) OVER() AS total_count,
		m.id,
		m.name,
		m.order_num,
		m.course_id,
		c.name,
		m.created_at,
		m.updated_at
	FROM
		modules m
	LEFT JOIN
		courses c ON m.course_id = c.id
	WHERE
		m.deleted_at = 0
	`

	var args []interface{}

	if req.Name != "" && req.Name != "string" {
		query += fmt.Sprintf(" AND m.name::text ILIKE $%d", len(args)+1)
		args = append(args, "%"+req.Name+"%")
	}
	if req.CourseId != "" && req.CourseId != "string" {
		query += fmt.Sprintf(" AND m.course_id = $%d", len(args)+1)
		args = append(args, req.CourseId)
	}

	query += " ORDER BY m.order_num ASC"
	args = append(args, req.Filter.Limit, req.Filter.Page)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting modules: %w", err)
	}
	defer rows.Close()

	modules := m.ModuleList{}
	for rows.Next() {
		module := m.ModuleRes{}
		var nameJSON []byte
		var courseNameJSON []byte
		var totalCount int

		err := rows.Scan(
			&totalCount,
			&module.Id,
			&nameJSON,
			&module.OrderNum,
			&module.CourseId,
			&courseNameJSON,
			&module.CreatedAt,
			&module.UpdatedAt,
		)

		json.Unmarshal(nameJSON, &module.Name)
		json.Unmarshal(courseNameJSON, &module.CourseName)

		if err != nil {
			return nil, nil, fmt.Errorf("error while scanning course row: %w", err)
		}
		modules.Modules = append(modules.Modules, &module)
		modules.Count = totalCount
	}
	response := m.NewPaginationResponse(req.Filter, modules.Count, modules.Modules)
	return &response, &modules, nil
}

func (r *ModuleRepo) Update(req *m.ModuleUpdateReq) (*m.Void, error) {
	query := `
	UPDATE modules
	SET
	`

	var conditions []string
	var args []interface{}

	if req.Name != "" && req.Name != "string" {
		conditions = append(conditions, " name = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.Name)
	}

	conditions = append(conditions, " course_id = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.CourseId)

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
		return nil, fmt.Errorf("error while updating module: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("module with id %s couldnt update", req.Id)
	}
	return &m.Void{}, nil
}

func (r *ModuleRepo) Delete(req *m.ById) (*m.Void, error) {
	query := `
	UPDATE modules
	SET
		deleted_at = EXTRACT(EPOCH FROM NOW())
	WHERE
		id = $1
	AND
		deleted_at = 0
	`

	res, err := r.db.Exec(query, req.Id)
	if err != nil {
		return nil, fmt.Errorf("error while deleting module: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("module with id %s couldnt delete", req.Id)
	}

	return &m.Void{}, nil
}

// func (r ModuleRepo) GetByIdMobile(req *m.ById) (*m.ModuleMobileRes, error) {
// 	return nil, nil
// }
// func (r ModuleRepo) GetAllMobile(req *m.ModuleGetListMobileReq) (*m.PaginationResponse, *m.ModuleMobileList, error) {
// 	return nil, nil, nil
// }
