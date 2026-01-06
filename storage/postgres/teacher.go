package postgres

import (
	"database/sql"
	"fmt"
	m "imed/pkg/models"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TeacherRepo struct {
	db *sql.DB
}

func NewTeacherRepo(db *sql.DB) *TeacherRepo {
	return &TeacherRepo{
		db: db,
	}
}

func (r *TeacherRepo) Create(req *m.TeacherCreateReq) (*m.ById, error) {
	id := uuid.NewString()

	if req.Password == "" {
		password := req.PhoneNumber
		req.Password = password
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return &m.ById{}, fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
	INSERT INTO teachers (
		id,
		name,
		phone_number,
		password,
		login
	) VALUES ($1, $2, $3, $4, $5)
	`

	_, err = r.db.Exec(query,
		id,
		req.Name,
		req.PhoneNumber,
		hashedPassword,
		req.Login,
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating teacher: %w", err)
	}

	return &m.ById{Id: id}, nil
}
func (r *TeacherRepo) GetById(req *m.ById) (*m.TeacherRes, error) {
	query := `
	SELECT 
		id,
		name,
		phone_number,
		login,
		created_at,
		updated_at
	FROM
		teachers
	WHERE
		deleted_at = 0
	AND
		id = $1
	`

	row := r.db.QueryRow(query, req.Id)

	teacher := m.TeacherRes{}
	err := row.Scan(
		&teacher.Id,
		&teacher.Name,
		&teacher.PhoneNumber,
		&teacher.Login,
		&teacher.CreatedAt,
		&teacher.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("teacher not found with id: %s", req.Id)
	}

	if err != nil {
		return nil, fmt.Errorf("error while getting teacher by id: %w", err)
	}

	return &teacher, nil
}
func (r *TeacherRepo) GetAll(req *m.TeacherGetListReq) (*m.PaginationResponse, *m.TeacherList, error) {
	query := `
	SELECT 
		COUNT(id) OVER() AS total_count,
		id,
		name,
		phone_number,
		login,
		created_at,
		updated_at
	FROM
		teachers
	WHERE
		deleted_at = 0
	`

	var args []interface{}
	if req.Name != "" && req.Name != "string" {
		query += fmt.Sprintf(" AND name ILIKE $%d", len(args)+1)
		args = append(args, "%"+req.Name+"%")
	}

	if req.PhoneNumber != "" && req.PhoneNumber != "string" {
		query += fmt.Sprintf(" AND phone_number ILIKE $%d", len(args)+1)
		args = append(args, "%"+req.PhoneNumber+"%")
	}
	query += " ORDER BY created_at DESC"

	args = append(args, req.Filter.Limit, req.Filter.Page)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting teacher: %w", err)
	}
	defer rows.Close()

	teachers := m.TeacherList{}

	for rows.Next() {
		var totalCount int
		teacher := m.TeacherRes{}
		err := rows.Scan(
			&totalCount,
			&teacher.Id,
			&teacher.Name,
			&teacher.PhoneNumber,
			&teacher.Login,
			&teacher.CreatedAt,
			&teacher.UpdatedAt,
		)

		if err != nil {
			return nil, nil, fmt.Errorf("error while scanning teacher row: %w", err)
		}

		teachers.Teachers = append(teachers.Teachers, &teacher)
		teachers.Count = totalCount
	}

	response := m.NewPaginationResponse(req.Filter, teachers.Count, teachers.Teachers)

	return &response, &teachers, nil
}
func (r *TeacherRepo) Update(req *m.TeacherUpdateReq) (*m.Void, error) {
	query := `
	UPDATE teachers
	SET
	`

	var conditions []string
	var args []interface{}

	if req.Name != "" && req.Name != "string" {
		conditions = append(conditions, " name = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.Name)
	}
	if req.PhoneNumber != "" && req.PhoneNumber != "string" {
		conditions = append(conditions, " phone_number = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.PhoneNumber)
	}

	if len(conditions) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	conditions = append(conditions, " updated_at = now()")
	query += strings.Join(conditions, ", ")
	query += " WHERE id = $" + strconv.Itoa(len(args)+1) + " AND deleted_at = 0"

	args = append(args, req.Id)

	res, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error while updating teacher: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("teacher with id %s couldnt update", req.Id)
	}
	return &m.Void{}, nil
}
func (r *TeacherRepo) Delete(req *m.ById) (*m.Void, error) {
	query := `
	UPDATE teachers
	SET
		deleted_at = EXTRACT(EPOCH FROM NOW())
	WHERE
		id = $1
	AND
		deleted_at = 0
	`

	res, err := r.db.Exec(query, req.Id)
	if err != nil {
		return nil, fmt.Errorf("error while deleting teacher: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("teacher with id %s couldnt delete", req.Id)
	}

	return &m.Void{}, nil
}
