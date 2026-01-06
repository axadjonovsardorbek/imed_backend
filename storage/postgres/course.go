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

type CourseRepo struct {
	db *sql.DB
}

func NewCourseRepo(db *sql.DB) *CourseRepo {
	return &CourseRepo{
		db: db,
	}
}

func (r *CourseRepo) Create(req *m.CourseCreateReq) (*m.ById, error) {
	id := uuid.NewString()
	query := `
	INSERT INTO courses (
		id,
		name,
		description,
		subject_id,
		teacher_id,
		image_url,
		price,
		is_public,
		order_num
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(query,
		id,
		req.Name,
		req.Description,
		req.SubjectId,
		req.TeacherId,
		req.ImageUrl,
		req.Price,
		req.IsPublic,
		req.OrderNum,
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating course: %w", err)
	}

	return &m.ById{Id: id}, nil
}

func (r *CourseRepo) GetById(req *m.ById) (*m.CourseRes, error) {

	query := `
	SELECT 
		c.id,
		c.name,
		c.description,
		c.price,
		c.image_url,
		c.is_public,
		c.order_num,
		c.subject_id,
		s.name,
		c.teacher_id,
		t.name,
		t.phone_number,
		c.created_at,
		c.updated_at
	FROM
		courses c
	LEFT JOIN
		subjects s ON c.subject_id = s.id
	LEFT JOIN
		teachers t ON c.teacher_id = t.id
	WHERE
		c.deleted_at = 0
	AND
		c.id = $1
	`

	row := r.db.QueryRow(query, req.Id)

	course := m.CourseRes{
		Subject: m.SubjectRes{},
		Teacher: m.TeacherRes{},
	}
	var nameJSON []byte
	var descriptionJSON []byte
	var subjectNameJSON []byte
	var priceJSON []byte

	err := row.Scan(
		&course.Id,
		&nameJSON,
		&descriptionJSON,
		&priceJSON,
		&course.ImageUrl,
		&course.IsPublic,
		&course.OrderNum,
		&course.Subject.Id,
		&subjectNameJSON,
		&course.Teacher.Id,
		&course.Teacher.Name,
		&course.Teacher.PhoneNumber,
		&course.CreatedAt,
		&course.UpdatedAt,
	)

	json.Unmarshal(nameJSON, &course.Name)
	json.Unmarshal(descriptionJSON, &course.Description)
	json.Unmarshal(subjectNameJSON, &course.Subject.Name)
	json.Unmarshal(priceJSON, &course.Price)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("course not found with id: %s", req.Id)
	}
	if err != nil {
		return nil, fmt.Errorf("error while getting course by id: %w", err)
	}
	return &course, nil
}

func (r *CourseRepo) GetAll(req *m.CourseGetListReq) (*m.PaginationResponse, *m.CourseList, error) {
	query := `
	SELECT 
		COUNT(c.id) OVER() AS total_count,
		c.id,
		c.name,
		c.description,
		c.price,
		c.image_url,
		c.is_public,
		c.order_num,
		c.subject_id,
		s.name,
		c.teacher_id,
		t.name,
		t.phone_number,
		c.created_at,
		c.updated_at
	FROM
		courses c
	LEFT JOIN
		subjects s ON c.subject_id = s.id
	LEFT JOIN
		teachers t ON c.teacher_id = t.id
	WHERE
		c.deleted_at = 0
	`

	var args []interface{}

	if req.Name != "" && req.Name != "string" {
		query += fmt.Sprintf(" AND c.name::text ILIKE $%d", len(args)+1)
		args = append(args, "%"+req.Name+"%")
	}
	if req.SubjectId != "" && req.SubjectId != "string" {
		query += fmt.Sprintf(" AND c.subject_id = $%d", len(args)+1)
		args = append(args, req.SubjectId)
	}
	if req.TeacherId != "" && req.TeacherId != "string" {
		query += fmt.Sprintf(" AND c.teacher_id = $%d", len(args)+1)
		args = append(args, req.TeacherId)
	}
	if req.IsPublic != "" && req.IsPublic != "string" {
		if req.IsPublic == "true" {
			query += fmt.Sprintf(" AND c.is_public = $%d", len(args)+1)
			args = append(args, true)
		} else if req.IsPublic == "false" {
			query += fmt.Sprintf(" AND c.is_public = $%d", len(args)+1)
			args = append(args, false)
		}
	}

	query += " ORDER BY c.order_num ASC"
	args = append(args, req.Filter.Limit, req.Filter.Page)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting courses: %w", err)
	}
	defer rows.Close()

	courses := m.CourseList{}
	for rows.Next() {
		course := m.CourseRes{
			Subject: m.SubjectRes{},
			Teacher: m.TeacherRes{},
		}
		var nameJSON []byte
		var descriptionJSON []byte
		var subjectNameJSON []byte
		var priceJSON []byte
		var totalCount int

		err := rows.Scan(
			&totalCount,
			&course.Id,
			&nameJSON,
			&descriptionJSON,
			&priceJSON,
			&course.ImageUrl,
			&course.IsPublic,
			&course.OrderNum,
			&course.Subject.Id,
			&subjectNameJSON,
			&course.Teacher.Id,
			&course.Teacher.Name,
			&course.Teacher.PhoneNumber,
			&course.CreatedAt,
			&course.UpdatedAt,
		)

		json.Unmarshal(nameJSON, &course.Name)
		json.Unmarshal(descriptionJSON, &course.Description)
		json.Unmarshal(subjectNameJSON, &course.Subject.Name)
		json.Unmarshal(priceJSON, &course.Price)

		if err != nil {
			return nil, nil, fmt.Errorf("error while scanning course row: %w", err)
		}
		courses.Courses = append(courses.Courses, &course)
		courses.Count = totalCount
	}
	response := m.NewPaginationResponse(req.Filter, courses.Count, courses.Courses)
	return &response, &courses, nil
}

func (r *CourseRepo) Update(req *m.CourseUpdateReq) (*m.Void, error) {
	query := `
	UPDATE courses
	SET
	`

	var conditions []string
	var args []interface{}

	if req.Name != "" && req.Name != "string" {
		conditions = append(conditions, " name = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.Name)
	}

	conditions = append(conditions, " description = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.Description)

	conditions = append(conditions, " subject_id = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.SubjectId)

	conditions = append(conditions, " teacher_id = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.TeacherId)

	conditions = append(conditions, " price = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.Price)

	conditions = append(conditions, " image_url = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.ImageUrl)

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
		return nil, fmt.Errorf("error while updating course: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("course with id %s couldnt update", req.Id)
	}
	return &m.Void{}, nil
}

func (r *CourseRepo) Delete(req *m.ById) (*m.Void, error) {
	query := `
	UPDATE courses
	SET
		deleted_at = EXTRACT(EPOCH FROM NOW())
	WHERE
		id = $1
	AND
		deleted_at = 0
	`

	res, err := r.db.Exec(query, req.Id)
	if err != nil {
		return nil, fmt.Errorf("error while deleting course: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("course with id %s couldnt delete", req.Id)
	}

	return &m.Void{}, nil
}

func (r *CourseRepo) BuyCourse(req *m.UserCourseCreateReq) (*m.Void, error) {

	id := uuid.NewString()
	query := `
	INSERT INTO user_courses (
		id,
		user_id,
		course_id,
		type,
		started_at,
		ended_at
	) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(query,
		id,
		req.UserId,
		req.CourseId,
		req.Type,
		req.StartedAt,
		req.EndedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error while buying course: %w", err)
	}

	return &m.Void{}, nil
}

func (r *CourseRepo) UserCourse(req *m.ById) (*m.UserCourseRes, error) {

	query := `
	SELECT 
		c.id,
		c.user_id,
		u.name,
		c.course_id,
		co.name,
		c.is_active,
		c.type,
		c.started_at,
		c.ended_at,
		c.created_at,
		c.updated_at
	FROM
		user_courses c
	LEFT JOIN
		users u ON c.user_id = u.id
	LEFT JOIN
		courses co ON c.course_id = co.id
	WHERE
		c.deleted_at = 0
	AND
		c.id = $1
	`

	row := r.db.QueryRow(query, req.Id)

	course := m.UserCourseRes{
		User:   &m.UserRes{},
		Course: &m.CourseRes{},
	}
	var courseNameJSON []byte

	err := row.Scan(
		&course.Id,
		&course.User.Id,
		&course.User.Name,
		&course.Course.Id,
		&courseNameJSON,
		&course.IsActive,
		&course.Type,
		&course.StartedAt,
		&course.EndedAt,
		&course.CreatedAt,
		&course.UpdatedAt,
	)

	json.Unmarshal(courseNameJSON, &course.Course.Name)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user course not found with id: %s", req.Id)
	}
	if err != nil {
		return nil, fmt.Errorf("error while getting user course by id: %w", err)
	}
	return &course, nil
}
func (r *CourseRepo) UserCourses(req *m.UserCourseGetListReq) (*m.PaginationResponse, *m.UserCourseList, error) {
	query := `
	SELECT 
		COUNT(c.id) OVER() AS total_count,
		c.id,
		c.user_id,
		u.name,
		c.course_id,
		co.name,
		c.is_active,
		c.type,
		c.started_at,
		c.ended_at,
		c.created_at,
		c.updated_at
	FROM
		user_courses c
	LEFT JOIN
		users u ON c.user_id = u.id
	LEFT JOIN
		courses co ON c.course_id = co.id
	WHERE
		c.deleted_at = 0
	`

	var args []interface{}

	if req.CourseId != "" && req.CourseId != "string" {
		query += fmt.Sprintf(" AND c.course_id = $%d", len(args)+1)
		args = append(args, req.CourseId)
	}
	if req.UserId != "" && req.UserId != "string" {
		query += fmt.Sprintf(" AND c.user_id = $%d", len(args)+1)
		args = append(args, req.UserId)
	}
	if req.Type != "" && req.Type != "string" {
		query += fmt.Sprintf(" AND c.type = $%d", len(args)+1)
		args = append(args, req.Type)
	}
	if req.IsActive != "" && req.IsActive != "string" {
		if req.IsActive == "true" {
			query += fmt.Sprintf(" AND c.is_active = $%d", len(args)+1)
			args = append(args, true)
		} else if req.IsActive == "false" {
			query += fmt.Sprintf(" AND c.is_active = $%d", len(args)+1)
			args = append(args, false)
		}
	}

	query += " ORDER BY c.created_at DESC"
	args = append(args, req.Filter.Limit, req.Filter.Page)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting user courses: %w", err)
	}
	defer rows.Close()

	courses := m.UserCourseList{}
	for rows.Next() {
		course := m.UserCourseRes{
			User:   &m.UserRes{},
			Course: &m.CourseRes{},
		}
		var courseNameJSON []byte
		var totalCount int

		err := rows.Scan(
			&totalCount,
			&course.Id,
			&course.User.Id,
			&course.User.Name,
			&course.Course.Id,
			&courseNameJSON,
			&course.IsActive,
			&course.Type,
			&course.StartedAt,
			&course.EndedAt,
			&course.CreatedAt,
			&course.UpdatedAt,
		)

		json.Unmarshal(courseNameJSON, &course.Course.Name)

		if err != nil {
			return nil, nil, fmt.Errorf("error while scanning user course row: %w", err)
		}
		courses.UserCourses = append(courses.UserCourses, &course)
		courses.Count = totalCount
	}
	response := m.NewPaginationResponse(req.Filter, courses.Count, courses.UserCourses)
	return &response, &courses, nil
}
func (r *CourseRepo) UserUpdateCourse(req *m.UserCourseUpdateReq) (*m.Void, error) {
	query := `
	UPDATE user_courses
	SET
	`

	var conditions []string
	var args []interface{}

	if req.UserId != "" && req.UserId != "string" {
		conditions = append(conditions, " user_id = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.UserId)
	}
	if req.CourseId != "" && req.CourseId != "string" {
		conditions = append(conditions, " course_id = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.CourseId)
	}

	conditions = append(conditions, " is_active = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.IsActive)

	conditions = append(conditions, " type = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.Type)

	conditions = append(conditions, " started_at = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.StartedAt)

	conditions = append(conditions, " ended_at = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.EndedAt)

	if len(conditions) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	conditions = append(conditions, " updated_at = now()")
	query += strings.Join(conditions, ", ")
	query += " WHERE id = $" + strconv.Itoa(len(args)+1) + " AND deleted_at = 0"

	args = append(args, req.Id)

	res, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error while updating user course: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("user course with id %s couldnt update", req.Id)
	}
	return &m.Void{}, nil
}
func (r *CourseRepo) DeleteUserCourse(req *m.ById) (*m.Void, error) {
	query := `
	UPDATE user_courses
	SET
		deleted_at = EXTRACT(EPOCH FROM NOW())
	WHERE
		id = $1
	AND
		deleted_at = 0
	`

	res, err := r.db.Exec(query, req.Id)
	if err != nil {
		return nil, fmt.Errorf("error while deleting user course: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("user course with id %s couldnt delete", req.Id)
	}

	return &m.Void{}, nil
}
