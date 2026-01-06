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

type SubjectRepo struct {
	db *sql.DB
}

func NewSubjectRepo(db *sql.DB) *SubjectRepo {
	return &SubjectRepo{
		db: db,
	}
}

func (r *SubjectRepo) Create(req *m.SubjectCreateReq) (*m.ById, error) {
	id := uuid.NewString()
	query := `
	INSERT INTO subjects (
		id,
		name,
		image_url,
		order_num
	) VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(query,
		id,
		req.Name,
		req.ImageUrl,
		req.OrderNum,
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating subject: %w", err)
	}

	return &m.ById{Id: id}, nil
}
func (r *SubjectRepo) GetById(req *m.ById) (*m.SubjectRes, error) {
	query := `
	SELECT 
		id,
		name,
		image_url,
		order_num,
		created_at,
		updated_at
	FROM
		subjects
	WHERE
		deleted_at = 0
	AND
		id = $1
	`

	row := r.db.QueryRow(query, req.Id)

	var nameJSON []byte

	subject := m.SubjectRes{}
	err := row.Scan(
		&subject.Id,
		&nameJSON,
		&subject.ImageUrl,
		&subject.OrderNum,
		&subject.CreatedAt,
		&subject.UpdatedAt,
	)

	json.Unmarshal(nameJSON, &subject.Name)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("subject not found with id: %s", req.Id)
	}

	if err != nil {
		return nil, fmt.Errorf("error while getting subject by id: %w", err)
	}

	return &subject, nil
}
func (r *SubjectRepo) GetAll(req *m.SubjectGetListReq) (*m.PaginationResponse, *m.SubjectList, error) {
	query := `
	SELECT 
		COUNT(id) OVER() AS total_count,
		id,
		name,
		image_url,
		order_num,
		created_at,
		updated_at
	FROM
		subjects
	WHERE
		deleted_at = 0
	`

	var args []interface{}
	if req.Name != "" && req.Name != "string" {
		query += fmt.Sprintf(" AND name::text ILIKE $%d", len(args)+1)
		args = append(args, "%"+req.Name+"%")
	}

	query += " ORDER BY order_num ASC"

	args = append(args, req.Filter.Limit, req.Filter.Page)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting subject: %w", err)
	}
	defer rows.Close()

	subjects := m.SubjectList{}

	for rows.Next() {

		subject := m.SubjectRes{}
		var nameJSON []byte
		var totalCount int

		err := rows.Scan(
			&totalCount,
			&subject.Id,
			&nameJSON,
			&subject.ImageUrl,
			&subject.OrderNum,
			&subject.CreatedAt,
			&subject.UpdatedAt,
		)

		json.Unmarshal(nameJSON, &subject.Name)

		if err != nil {
			return nil, nil, fmt.Errorf("error while scanning subject row: %w", err)
		}

		subjects.Subjects = append(subjects.Subjects, &subject)
		subjects.Count = totalCount
	}

	response := m.NewPaginationResponse(req.Filter, subjects.Count, subjects.Subjects)

	return &response, &subjects, nil
}
func (r *SubjectRepo) Update(req *m.SubjectUpdateReq) (*m.Void, error) {
	query := `
	UPDATE subjects
	SET
	`

	var conditions []string
	var args []interface{}

	if req.Name != "" && req.Name != "string" {
		conditions = append(conditions, " name = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.Name)
	}

	conditions = append(conditions, " image_url = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.ImageUrl)

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
		return nil, fmt.Errorf("error while updating subject: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("subject with id %s couldnt update", req.Id)
	}
	return &m.Void{}, nil
}
func (r *SubjectRepo) Delete(req *m.ById) (*m.Void, error) {
	query := `
	UPDATE subjects
	SET
		deleted_at = EXTRACT(EPOCH FROM NOW())
	WHERE
		id = $1
	AND
		deleted_at = 0
	`

	res, err := r.db.Exec(query, req.Id)
	if err != nil {
		return nil, fmt.Errorf("error while deleting subject: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("subject with id %s couldnt delete", req.Id)
	}

	return &m.Void{}, nil
}
