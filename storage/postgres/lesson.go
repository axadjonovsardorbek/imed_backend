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
		order_num,
		is_public
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(query,
		id,
		req.Name,
		req.ModuleId,
		req.Type,
		req.Duration,
		req.OrderNum,
		req.IsPublic,
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating lesson: %w", err)
	}

	return &m.ById{Id: id}, nil
}

func (r *LessonRepo) GetLessonCourseInfo(req *m.ById) (*m.LessonCourseInfo, error) {
	var args []interface{}

	query := `
	SELECT
	  c.id AS course_id,
	  (c.name->>$1) AS course_name,
	  l.id AS lesson_id,
	  (l.name->>$1) AS lesson_name
	FROM lessons l
	JOIN modules m ON m.id = l.module_id AND m.deleted_at=0
	JOIN courses c ON c.id = m.course_id AND c.deleted_at=0
	WHERE l.id=$2 AND l.deleted_at=0;
	`
	args = append(args, "uz", req.Id)

	row := r.db.QueryRow(query, args...)

	lesson := m.LessonCourseInfo{}

	err := row.Scan(
		&lesson.CourseId,
		&lesson.CourseName,
		&lesson.LessonId,
		&lesson.LessonName,
	)

	if err != nil {
		return nil, fmt.Errorf("error while getting lesson by id: %w", err)
	}
	return &lesson, nil
}

func (r *LessonRepo) GetById(req *m.ById) (*m.LessonRes, error) {
	var args []interface{}

	query := `
		SELECT 
			l.id,
			l.name,
			l.type,
			l.duration,
			l.order_num,
			l.module_id,
			m.name,
			l.is_public,
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
		&lesson.IsPublic,
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
	var args []interface{}

	query := `
		SELECT 
			COUNT(l.id) OVER() AS total_count,
			l.id,
			l.name,
			l.type,
			l.duration,
			l.order_num,
			l.module_id,
			m.name,
			l.is_public,
			l.created_at,
			l.updated_at
		FROM
			lessons l
		LEFT JOIN
			modules m ON l.module_id = m.id
		WHERE
			l.deleted_at = 0
		`

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
	if req.IsPublic != "" && req.IsPublic != "string" {
		if req.IsPublic == "true" {
			query += fmt.Sprintf(" AND l.is_public = TRUE")
		} else if req.IsPublic == "false" {
			query += fmt.Sprintf(" AND l.is_public = FALSE")
		}
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
			&lesson.IsPublic,
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

	conditions = append(conditions, " is_public = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.IsPublic)

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

func (r *LessonRepo) GetByIdMobile(req *m.ByIdMobile) (*m.SourceLessonMobileRes, error) {
	var args []interface{}

	query := `
	WITH l AS (
		SELECT
			le.id,
			le.name ->> $3        AS name,
			le.duration::int      AS duration,
			le.order_num::int     AS order_num,
			le.type AS type
		FROM lessons le
		WHERE le.deleted_at = 0
		  AND le.id = $1
	),
	completed AS (
		SELECT
			EXISTS (
				SELECT 1
				FROM user_lessons ul
				WHERE ul.lesson_id = $1
				  AND ul.user_id = $2
				  AND ul.is_completed = TRUE
			) AS is_completed
	),
	src AS (
		SELECT
			s.id::text            AS id,
			(s.name ->> $3)       AS name,
			(s.url  ->> $3)       AS url,
			s.type                AS type,
			s.order_num::int      AS order_num
		FROM sources s
		WHERE s.deleted_at = 0
		  AND s.lesson_id = $1
	)
	SELECT
		l.id::text                           AS id,
		l.name                               AS name,
		l.type                               AS type,
		l.duration                           AS duration,
		l.order_num                          AS order_num,
		c.is_completed                       AS is_completed,
	
		COALESCE((
			SELECT jsonb_agg(
				jsonb_build_object(
					'id', id,
					'name', name,
					'url', url,
					'type', type,
					'order_num', order_num
				)
				ORDER BY order_num
			)
			FROM src
			WHERE type = 'test'
		), '[]'::jsonb) AS tests,
	
		COALESCE((
			SELECT jsonb_agg(
				jsonb_build_object(
					'id', id,
					'name', name,
					'url', url,
					'type', type,
					'order_num', order_num
				)
				ORDER BY order_num
			)
			FROM src
			WHERE type = 'document'
		), '[]'::jsonb) AS documents,
	
		COALESCE((
			SELECT jsonb_agg(
				jsonb_build_object(
					'id', id,
					'name', name,
					'url', url,
					'video_url', video_url,
					'type', type,
					'order_num', order_num
				)
				ORDER BY order_num
			)
			FROM src
			WHERE type = 'video'
		), '[]'::jsonb) AS videos
	
	FROM l
	CROSS JOIN completed c
	`

	args = append(args, req.Id, req.UserId, req.Language)

	row := r.db.QueryRow(query, args...)

	res := m.SourceLessonMobileRes{}

	var (
		testsRaw     []byte
		documentsRaw []byte
		videosRaw    []byte
	)

	if err := row.Scan(
		&res.Id,
		&res.Name,
		&res.Type,
		&res.Duration,
		&res.OrderNum,
		&res.IsCompleted,
		&testsRaw,
		&documentsRaw,
		&videosRaw,
	); err != nil {
		return nil, err
	}

	res.Tests = make([]m.SourceMobile, 0)
	res.Documents = make([]m.SourceMobile, 0)
	res.Videos = make([]m.SourceMobile, 0)

	if len(testsRaw) > 0 {
		_ = json.Unmarshal(testsRaw, &res.Tests)
	}
	if len(documentsRaw) > 0 {
		_ = json.Unmarshal(documentsRaw, &res.Documents)
	}
	if len(videosRaw) > 0 {
		_ = json.Unmarshal(videosRaw, &res.Videos)
	}

	return &res, nil
}
func (r *LessonRepo) GetAllMobile(req *m.LessonGetListMobileReq) (*m.PaginationResponse, *m.LessonMobileList, error) {
	return nil, nil, nil
}
