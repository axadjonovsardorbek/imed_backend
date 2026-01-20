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

type LessonRepo struct {
	db *sql.DB
}

func NewLessonRepo(db *sql.DB) *LessonRepo {
	return &LessonRepo{
		db: db,
	}
}

func (r *LessonRepo) Create(req *m.LessonCreateReq) (*m.ById, error) {
	id := uuid.NewString()
	query := `
	INSERT INTO lessons (
		id,
		name,
		module_id,
		type,
		duration,
		order_num
	) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(query,
		id,
		req.Name,
		req.ModuleId,
		req.Type,
		req.Duration,
		req.OrderNum,
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating lesson: %w", err)
	}

	return &m.ById{Id: id}, nil
}

func (r *LessonRepo) GetById(req *m.CourseById) (*m.LessonRes, error) {
	var query string
	var args []interface{}

	if req.Role != "user" || req.UserId == "" {
		query = `
		SELECT 
			l.id,
			l.name,
			l.type,
			l.duration,
			l.order_num,
			l.module_id,
			m.name,
			1 AS is_completed,
			l.created_at,
			l.updated_at
		FROM
			lessons l
		LEFT JOIN
			modules m ON l.module_id = m.id
		WHERE
			l.deleted_at = 0
		AND 
			l.id = $1
		`
		args = append(args, req.Id)
	} else {
		query = `
		SELECT 
			l.id,
			l.name,
			l.type,
			l.duration,
			l.order_num,
			l.module_id,
			m.name,
			COALESCE(ul.is_completed, false) AS is_completed,
			l.created_at,
			l.updated_at
		FROM
			lessons l
		LEFT JOIN
			user_lessons ul
		ON 
			ul.lesson_id = l.id
		AND
			ul.user_id = $1
		LEFT JOIN
			modules m ON l.module_id = m.id
		WHERE
			l.deleted_at = 0
		AND
			l.id = $2
		`
		args = append(args, req.UserId, req.Id)
	}

	row := r.db.QueryRow(query, args...)

	lesson := m.LessonRes{}
	var nameJSON []byte
	var moduleNameJSON []byte

	err := row.Scan(
		&lesson.Id,
		&nameJSON,
		&lesson.Type,
		&lesson.Duration,
		&lesson.OrderNum,
		&lesson.ModuleId,
		&moduleNameJSON,
		&lesson.IsCompleted,
		&lesson.CreatedAt,
		&lesson.UpdatedAt,
	)

	json.Unmarshal(nameJSON, &lesson.Name)
	json.Unmarshal(moduleNameJSON, &lesson.ModuleName)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lesson not found with id: %s", req.Id)
	}
	if err != nil {
		return nil, fmt.Errorf("error while getting lesson by id: %w", err)
	}
	return &lesson, nil
}

func (r *LessonRepo) GetAll(req *m.LessonGetListReq) (*m.PaginationResponse, *m.LessonList, error) {
	var query string
	var args []interface{}

	if req.Role != "user" || req.UserId == "" {
		query = `
		SELECT 
			COUNT(l.id) OVER() AS total_count,
			l.id,
			l.name,
			l.type,
			l.duration,
			l.order_num,
			l.module_id,
			m.name,
			1 AS is_completed,
			l.created_at,
			l.updated_at
		FROM
			lessons l
		LEFT JOIN
			modules m ON l.module_id = m.id
		WHERE
			l.deleted_at = 0
		`
	} else {
		query = `
		SELECT 
			COUNT(l.id) OVER() AS total_count,
			l.id,
			l.name,
			l.type,
			l.duration,
			l.order_num,
			l.module_id,
			m.name,
			COALESCE(ul.is_completed, false) AS is_completed,
			l.created_at,
			l.updated_at
		FROM
			lessons l
		LEFT JOIN
			user_lessons ul
		ON 
			ul.lesson_id = l.id
		AND
			ul.user_id = $1
		LEFT JOIN
			modules m ON l.module_id = m.id
		WHERE
			l.deleted_at = 0
		`
		args = append(args, req.UserId)
	}

	if req.Name != "" && req.Name != "string" {
		query += fmt.Sprintf(" AND l.name::text ILIKE $%d", len(args)+1)
		args = append(args, "%"+req.Name+"%")
	}
	if req.ModuleId != "" && req.ModuleId != "string" {
		query += fmt.Sprintf(" AND l.module_id = $%d", len(args)+1)
		args = append(args, req.ModuleId)
	}
	if req.Type != "" && req.Type != "string" {
		query += fmt.Sprintf(" AND l.type = $%d", len(args)+1)
		args = append(args, req.Type)
	}

	query += " ORDER BY l.order_num ASC"
	args = append(args, req.Filter.Limit, req.Filter.Page)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting lessons: %w", err)
	}
	defer rows.Close()

	lessons := m.LessonList{}
	for rows.Next() {
		lesson := m.LessonRes{}
		var nameJSON []byte
		var moduleNameJSON []byte
		var totalCount int

		err := rows.Scan(
			&totalCount,
			&lesson.Id,
			&nameJSON,
			&lesson.Type,
			&lesson.Duration,
			&lesson.OrderNum,
			&lesson.ModuleId,
			&moduleNameJSON,
			&lesson.IsCompleted,
			&lesson.CreatedAt,
			&lesson.UpdatedAt,
		)

		json.Unmarshal(nameJSON, &lesson.Name)
		json.Unmarshal(moduleNameJSON, &lesson.ModuleName)

		if err != nil {
			return nil, nil, fmt.Errorf("error while scanning lesson row: %w", err)
		}
		lessons.Lessons = append(lessons.Lessons, &lesson)
		lessons.Count = totalCount
	}
	response := m.NewPaginationResponse(req.Filter, lessons.Count, lessons.Lessons)
	return &response, &lessons, nil
}

func (r *LessonRepo) Update(req *m.LessonUpdateReq) (*m.Void, error) {
	query := `
	UPDATE lessons
	SET
	`

	var conditions []string
	var args []interface{}

	if req.Name != "" && req.Name != "string" {
		conditions = append(conditions, " name = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.Name)
	}

	conditions = append(conditions, " module_id = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.ModuleId)

	conditions = append(conditions, " type = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.Type)

	conditions = append(conditions, " duration = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.Duration)

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
		return nil, fmt.Errorf("error while updating lesson: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("lesson with id %s couldnt update", req.Id)
	}
	return &m.Void{}, nil
}

func (r *LessonRepo) Delete(req *m.ById) (*m.Void, error) {
	query := `
	UPDATE lessons
	SET
		deleted_at = EXTRACT(EPOCH FROM NOW())
	WHERE
		id = $1
	AND
		deleted_at = 0
	`

	res, err := r.db.Exec(query, req.Id)
	if err != nil {
		return nil, fmt.Errorf("error while deleting lesson: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("lesson with id %s couldnt delete", req.Id)
	}

	return &m.Void{}, nil
}

func (r *LessonRepo) EndLesson(req *m.EndUserLessonReq) (*m.Void, error) {
	id := uuid.NewString()
	query := `
	INSERT INTO user_lessons (
		id,
		user_id,
		lesson_id
	) VALUES ($1, $2, $3)
	ON CONFLICT (user_id, lesson_id) DO NOTHING
	`

	_, err := r.db.Exec(query,
		id,
		req.UserId,
		req.LessonId,
	)

	if err != nil {
		return nil, fmt.Errorf("error while ending lesson: %w", err)
	}

	return &m.Void{}, nil
}
