package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
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
		order_num,
		can_buy
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
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
		req.CanBuy,
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating course: %w", err)
	}

	return &m.ById{Id: id}, nil
}

func (r *CourseRepo) GetById(req *m.ById) (*m.CourseRes, error) {
	var args []interface{}

	query := `
		SELECT 
			c.id,
			c.name,
			c.description,
			c.price,
			c.image_url,
			c.is_public,
			c.order_num,
			c.can_buy,
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
	args = append(args, req.Id)

	row := r.db.QueryRow(query, args...)

	course := m.CourseRes{}
	var nameJSON []byte
	var descriptionJSON []byte
	var subjectNameJSON []byte

	err := row.Scan(
		&course.Id,
		&nameJSON,
		&descriptionJSON,
		&course.Price,
		&course.ImageUrl,
		&course.IsPublic,
		&course.OrderNum,
		&course.CanBuy,
		&course.SubjectId,
		&subjectNameJSON,
		&course.TeacherId,
		&course.TeacherName,
		&course.TeacherPhoneNumber,
		&course.CreatedAt,
		&course.UpdatedAt,
	)

	json.Unmarshal(nameJSON, &course.Name)
	json.Unmarshal(descriptionJSON, &course.Description)
	json.Unmarshal(subjectNameJSON, &course.SubjectName)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("course not found with id: %s", req.Id)
	}
	if err != nil {
		return nil, fmt.Errorf("error while getting course by id: %w", err)
	}
	return &course, nil
}

func (r *CourseRepo) GetAll(req *m.CourseGetListReq) (*m.PaginationResponse, *m.CourseList, error) {

	var args []interface{}

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
			c.can_buy,
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
		course := m.CourseRes{}
		var nameJSON []byte
		var descriptionJSON []byte
		var subjectNameJSON []byte
		var totalCount int

		err := rows.Scan(
			&totalCount,
			&course.Id,
			&nameJSON,
			&descriptionJSON,
			&course.Price,
			&course.ImageUrl,
			&course.IsPublic,
			&course.OrderNum,
			&course.CanBuy,
			&course.SubjectId,
			&subjectNameJSON,
			&course.TeacherId,
			&course.TeacherName,
			&course.TeacherPhoneNumber,
			&course.CreatedAt,
			&course.UpdatedAt,
		)

		json.Unmarshal(nameJSON, &course.Name)
		json.Unmarshal(descriptionJSON, &course.Description)
		json.Unmarshal(subjectNameJSON, &course.SubjectName)

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

	if req.Description != "" && req.Description != "string" {
		conditions = append(conditions, " description = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.Description)
	}

	if req.SubjectId != "" && req.SubjectId != "string" {
		conditions = append(conditions, " subject_id = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.SubjectId)
	}

	if req.TeacherId != "" && req.TeacherId != "string" {
		conditions = append(conditions, " teacher_id = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.TeacherId)
	}

	if req.Price != nil && len(req.Price) > 0 {
		conditions = append(conditions, "price = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.Price)
	}

	if req.ImageUrl != "" && req.ImageUrl != "string" {
		conditions = append(conditions, " image_url = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.ImageUrl)
	}

	if req.OrderNum > 0 {
		conditions = append(conditions, " order_num = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.OrderNum)
	}

	conditions = append(conditions, " is_public = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.IsPublic)

	conditions = append(conditions, " can_buy = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.CanBuy)

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

	tz, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if p := recover(); p != nil {
			tz.Rollback()
			err = fmt.Errorf("panic occurred: %v", p)
		} else if err != nil {
			tz.Rollback()
		} else {
			err = tz.Commit()
		}
	}()

	var duration, price int

	getTariffQuery := `
	SELECT
	  t.duration,
	  (opt->>'price')::bigint AS price
	FROM tariffs t
	JOIN courses c ON c.id = $1
	CROSS JOIN LATERAL jsonb_array_elements(c.price) AS opt
	WHERE t.id = $2
	  AND t.deleted_at = 0
	  AND c.deleted_at = 0
	  AND (opt->>'duration')::bigint = t.duration
	LIMIT 1
	`

	err = tz.QueryRow(getTariffQuery, req.CourseId, req.TariffId).Scan(&duration, &price)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("price option not found for course=%s tariff=%s", req.CourseId, req.TariffId)
		}
		return nil, err
	}

	// is_free = true  -> summa 0
	// is_free = false va amount <= 0 -> kursning haqiqiy narxi
	// is_free = false va amount > 0  -> admin kiritgan summa
	var paidAmount float64
	switch {
	case req.IsFree:
		paidAmount = 0
	case req.Amount > 0:
		paidAmount = req.Amount
	default:
		paidAmount = float64(price)
	}

	comment := strings.TrimSpace(req.Comment)
	checkURL := strings.TrimSpace(req.CheckUrl)

	order_query := `
	INSERT INTO orders (
		id,
		user_id,
		course_id,
		tariff_id,
		amount,
		status,
		duration,
		course_amount,
		payment_type,
		comment,
		check_url,
		paid_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now())
	`

	order_id := uuid.NewString()
	_, err = tz.Exec(order_query,
		order_id,
		req.UserId,
		req.CourseId,
		req.TariffId,
		paidAmount,
		"PAID",
		duration,
		price,
		"admin",
		comment,
		checkURL,
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating order: %w", err)
	}

	id := uuid.NewString()
	query := `
	INSERT INTO user_courses (
		id,
		user_id,
		course_id,
		tariff_id,
		started_at,
		ended_at
	) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = tz.Exec(query,
		id,
		req.UserId,
		req.CourseId,
		req.TariffId,
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
		(
			c.started_at <= NOW()
			AND (c.ended_at IS NULL OR c.ended_at > NOW())
		) AS is_active,
		c.tariff_id,
		t.name,
		t.duration,
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
	LEFT JOIN
		tariffs t ON c.tariff_id = t.id
	WHERE
		c.deleted_at = 0
	AND
		c.id = $1
	`

	row := r.db.QueryRow(query, req.Id)

	course := m.UserCourseRes{}
	var courseNameJSON []byte

	err := row.Scan(
		&course.Id,
		&course.UserId,
		&course.UserName,
		&course.CourseId,
		&courseNameJSON,
		&course.IsActive,
		&course.Tariffid,
		&course.TariffName,
		&course.Duration,
		&course.StartedAt,
		&course.EndedAt,
		&course.CreatedAt,
		&course.UpdatedAt,
	)

	json.Unmarshal(courseNameJSON, &course.CourseName)

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
		(
			c.started_at <= NOW()
			AND (c.ended_at IS NULL OR c.ended_at > NOW())
		) AS is_active,
		c.tariff_id,
		t.name,
		t.duration,
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
	LEFT JOIN
		tariffs t ON c.tariff_id = t.id
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
	if req.TariffId != "" && req.TariffId != "string" {
		query += fmt.Sprintf(" AND c.tariff_id = $%d", len(args)+1)
		args = append(args, req.TariffId)
	}
	if req.IsActive != "" && req.IsActive != "string" {
		if req.IsActive == "true" {
			query += `
				AND c.started_at <= NOW()
				AND (c.ended_at IS NULL OR c.ended_at > NOW())
			`
		} else if req.IsActive == "false" {
			query += `
				AND (
					c.started_at > NOW()
					OR (c.ended_at IS NOT NULL AND c.ended_at <= NOW())
				)
			`
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
		course := m.UserCourseRes{}
		var courseNameJSON []byte
		var totalCount int

		err := rows.Scan(
			&totalCount,
			&course.Id,
			&course.UserId,
			&course.UserName,
			&course.CourseId,
			&courseNameJSON,
			&course.IsActive,
			&course.Tariffid,
			&course.TariffName,
			&course.Duration,
			&course.StartedAt,
			&course.EndedAt,
			&course.CreatedAt,
			&course.UpdatedAt,
		)

		json.Unmarshal(courseNameJSON, &course.CourseName)

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

	if req.TariffId != "" && req.TariffId != "string" {
		conditions = append(conditions, " tariff_id = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.TariffId)
	}

	if req.StartedAt != "" && req.StartedAt != "string" {
		conditions = append(conditions, " started_at = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.StartedAt)
	}
	if req.EndedAt != "" && req.EndedAt != "string" {
		conditions = append(conditions, " ended_at = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.EndedAt)

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

func (r *CourseRepo) GetByIdMobile(req *m.ByIdMobile) (*m.MobileCourseRes, error) {

	query := `
	WITH course_base AS (
		SELECT
		  c.id,
		  c.teacher_id,
		  t.name AS teacher_name,
		  (c.name->>$1)        AS name,
		  (c.description->>$1) AS description,
		  c.image_url          AS image_url,
		  COALESCE(
		    (
		      SELECT jsonb_agg(elem ORDER BY (elem->>'duration')::int)
		      FROM jsonb_array_elements(COALESCE(c.price, '[]'::jsonb)) AS elem
		    ),
		    '[]'::jsonb
		  ) AS price,
		  c.is_public 		   AS is_public,
		  c.can_buy            AS can_buy,
	  
		  c.total_lessons      AS lessons,
		  c.total_sources      AS total_sources,
		  (
			SELECT COUNT(*)
			FROM tests t
			JOIN lessons l ON l.id = t.lesson_id
			JOIN modules md ON md.id = l.module_id
			WHERE md.course_id = c.id
			  AND t.deleted_at = 0 AND l.deleted_at = 0 AND md.deleted_at = 0
			  AND EXISTS (SELECT 1 FROM test_questions q WHERE q.test_id = t.id AND q.deleted_at = 0)
		  )                    AS tests,
		  c.total_documents    AS documents,
		  c.total_videos       AS videos,
		  c.total_duration     AS duration,
	  
		  (
			EXISTS (
			  SELECT 1
			  FROM user_courses uc
			  WHERE uc.user_id = $2
			    AND uc.course_id = c.id
			    AND uc.deleted_at = 0
			    AND uc.started_at <= NOW()
			    AND (uc.ended_at IS NULL OR uc.ended_at > NOW())
			)
			-- T-12: faol premium barcha kurslarni ochadi
			OR EXISTS (
			  SELECT 1 FROM user_subscriptions us
			  WHERE us.user_id = $2 AND us.started_at <= NOW() AND us.ended_at > NOW()
			)
			-- T-16: ustozga barcha kurslar doimiy ochiq
			OR EXISTS (
			  SELECT 1 FROM users u WHERE u.id = $2 AND u.is_ustoz
			)
		  ) AS has_access
		FROM courses c
		LEFT JOIN teachers t ON t.id = c.teacher_id
		WHERE c.deleted_at = 0
		  AND c.id = $3
	  ),
	  lessons_json AS (
		SELECT
		  l.module_id,
		  jsonb_agg(
			jsonb_build_object(
			  'id',        l.id::text,
			  'name',      (l.name->>$1),
			  'module_id', l.module_id::text,
			  'order_num', l.order_num,
			  'is_public', CASE
							 WHEN cb.has_access THEN true
							 ELSE COALESCE(l.is_public, false)
						   END,
			'is_completed', COALESCE(ul.is_completed, false),
			'content_types', lc.content_types
			)
			ORDER BY l.order_num
		  ) AS lessons
		FROM lessons l
		JOIN modules m ON m.id = l.module_id
		CROSS JOIN course_base cb
		LEFT JOIN user_lessons ul
		  ON ul.lesson_id = l.id
	      AND ul.user_id = $2
		-- T-06: faqat kontenti bor darslar; eski PDF testlar hisobga olinmaydi
		CROSS JOIN LATERAL (
		  SELECT to_jsonb(ARRAY_REMOVE(ARRAY[
		    CASE WHEN EXISTS (SELECT 1 FROM sources s WHERE s.lesson_id = l.id AND s.deleted_at = 0 AND s.type = 'video') THEN 'video' END,
		    CASE WHEN EXISTS (SELECT 1 FROM sources s WHERE s.lesson_id = l.id AND s.deleted_at = 0 AND s.type = 'audio') THEN 'audio' END,
		    CASE WHEN EXISTS (SELECT 1 FROM sources s WHERE s.lesson_id = l.id AND s.deleted_at = 0 AND s.type = 'document') THEN 'document' END,
		    CASE WHEN EXISTS (
		      SELECT 1 FROM tests t
		      WHERE t.lesson_id = l.id AND t.deleted_at = 0
		        AND EXISTS (SELECT 1 FROM test_questions q WHERE q.test_id = t.id AND q.deleted_at = 0)
		    ) THEN 'test' END
		  ], NULL)) AS content_types
		) lc
		WHERE m.course_id = $3
		  AND m.deleted_at = 0
		  AND l.deleted_at = 0
		  AND lc.content_types <> '[]'::jsonb
		GROUP BY l.module_id
	  ),
	  modules_json AS (
		SELECT
		  m.course_id,
		  jsonb_agg(
			jsonb_build_object(
			  'id',        m.id::text,
			  'name',      (m.name->>$1),
			  'order_num', m.order_num,
			  'course_id', m.course_id::text,
			  'lessons',   COALESCE(lj.lessons, '[]'::jsonb)
			)
			ORDER BY m.order_num
		  ) AS modules
		FROM modules m
		JOIN lessons_json lj ON lj.module_id = m.id
		WHERE m.course_id = $3
		  AND m.deleted_at = 0
		GROUP BY m.course_id
	  )
	  SELECT
		cb.id,
		cb.name,
		cb.description,
		cb.teacher_id,
		cb.teacher_name,
		cb.image_url,
		cb.has_access,
		cb.price,
		cb.is_public,
		cb.can_buy,
	  
		cb.tests,
		cb.documents,
		cb.duration,
		cb.lessons,
	  
		COALESCE(mj.modules, '[]'::jsonb) AS modules
	  FROM course_base cb
	  LEFT JOIN modules_json mj ON mj.course_id = cb.id	  
	`

	row := r.db.QueryRow(query, req.Language, req.UserId, req.Id)

	var res m.MobileCourseRes
	var modulesRaw []byte

	if err := row.Scan(
		&res.Id,
		&res.Name,
		&res.Description,
		&res.TeacherId,
		&res.TeacherName,
		&res.ImageUrl,
		&res.HasAccess,
		&res.Price,
		&res.IsPublic,
		&res.CanBuy,
		&res.Tests,
		&res.Documents,
		&res.Duration,
		&res.Lessons,
		&modulesRaw,
	); err != nil {
		return nil, err
	}

	if res.HasAccess {
		res.CanBuy = false
	}

	if len(modulesRaw) > 0 {
		if err := json.Unmarshal(modulesRaw, &res.Modules); err != nil {
			return nil, err
		}
	} else {
		res.Modules = make([]m.MobileCourseModuleRes, 0)
	}

	return &res, nil
}
func (r *CourseRepo) GetAllMobile(req *m.CourseGetListMobileReq) (*m.PaginationResponse, *m.CourseMobileList, error) {
	var args []interface{}

	query := `
	SELECT 
		COUNT(c.id) OVER() AS total_count,
		c.id,
		c.name->> $1,
		c.description->> $1,
		c.image_url,
		c.is_public,
		c.order_num,
		c.teacher_id,
		t.name
	FROM
		courses c
	LEFT JOIN
		teachers t ON c.teacher_id = t.id
	WHERE
		c.deleted_at = 0
	`
	args = append(args, req.Language)

	if req.Name != "" && req.Name != "string" {
		query += fmt.Sprintf(" AND c.name::text ILIKE $%d", len(args)+1)
		args = append(args, "%"+req.Name+"%")
	}
	if req.SubjectId != "" && req.SubjectId != "string" {
		query += fmt.Sprintf(" AND c.subject_id = $%d", len(args)+1)
		args = append(args, req.SubjectId)
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

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting courses: %w", err)
	}
	defer rows.Close()

	courses := m.CourseMobileList{}
	for rows.Next() {
		course := m.CourseMobileRes{}
		var totalCount int

		err := rows.Scan(
			&totalCount,
			&course.Id,
			&course.Name,
			&course.Description,
			&course.ImageUrl,
			&course.IsPublic,
			&course.OrderNum,
			&course.TeacherId,
			&course.TeacherName,
		)

		if err != nil {
			return nil, nil, fmt.Errorf("error while scanning course row: %w", err)
		}
		courses.Courses = append(courses.Courses, &course)
		courses.Count = totalCount
	}

	response := m.NewPaginationResponse(m.Pagination{Limit: courses.Count}, courses.Count, courses.Courses)

	return &response, &courses, nil
}
func (r *CourseRepo) UserCourseMobile(req *m.ByIdMobile) (*m.UserCourseMobileRes, error) {

	query := `
	SELECT 
		c.id,
		c.user_id,
		u.name,
		c.course_id,
		co.name->> $1,
		co.image_url,
		(
			c.started_at <= NOW()
			AND (c.ended_at IS NULL OR c.ended_at > NOW())
		) AS is_active,
		c.tariff_id,
		t.name,
		t.duration,
		c.started_at,
		c.ended_at,
		c.created_at,
		c.updated_at,
	FROM
		user_courses c
	LEFT JOIN
		users u ON c.user_id = u.id
	LEFT JOIN
		courses co ON c.course_id = co.id
	LEFT JOIN
		tariffs t ON c.tariff_id = t.id
	WHERE
		c.deleted_at = 0
	AND
		c.id = $2
	`

	row := r.db.QueryRow(query, req.Language, req.Id)

	course := m.UserCourseMobileRes{}

	err := row.Scan(
		&course.Id,
		&course.UserId,
		&course.UserName,
		&course.CourseId,
		&course.CourseName,
		&course.CourseImageUrl,
		&course.IsActive,
		&course.TariffId,
		&course.TariffName,
		&course.Duration,
		&course.StartedAt,
		&course.EndedAt,
		&course.CreatedAt,
		&course.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user course not found with id: %s", req.Id)
	}
	if err != nil {
		return nil, fmt.Errorf("error while getting user course by id: %w", err)
	}
	return &course, nil
}
func (r *CourseRepo) UserCoursesMobile(req *m.UserCourseGetListMobileReq) (*m.PaginationResponse, *m.UserCourseMobileList, error) {
	query := `
	SELECT 
	  COUNT(c.id) OVER() AS total_count,
	  c.id,
	  c.user_id,
	  u.name AS user_name,

	  c.course_id,
	  (co.name->> $1) AS course_name,
	  co.image_url,

	  (c.started_at <= NOW() AND (c.ended_at IS NULL OR c.ended_at > NOW())) AS is_active,

	  c.tariff_id,
	  t.name,
	  t.duration,
	  c.started_at,
	  c.ended_at,
	  c.created_at,
	  c.updated_at,

	  COALESCE(co.total_lessons, 0)::int AS total_lessons,
	  COALESCE(ucp.completed_lessons, 0)::int AS completed_lessons,

	  CASE
	  WHEN COALESCE(co.total_lessons, 0) = 0 THEN 0
	  ELSE ROUND(
		(COALESCE(ucp.completed_lessons, 0)::numeric * 100)
		/ co.total_lessons,
		2
	  )
	END AS progress_percentage

	FROM user_courses c
	LEFT JOIN users u
	  ON c.user_id = u.id
	LEFT JOIN tariffs t
	  ON c.tariff_id = t.id
	LEFT JOIN courses co
	  ON c.course_id = co.id
	LEFT JOIN user_course_progress ucp
	  ON ucp.user_id = c.user_id
	 AND ucp.course_id = c.course_id

	WHERE c.deleted_at = 0
	`

	var args []interface{}
	args = append(args, req.Language)

	if req.UserId != "" && req.UserId != "string" {
		query += fmt.Sprintf(" AND c.user_id = $%d", len(args)+1)
		args = append(args, req.UserId)
	}
	if req.TariffId != "" && req.TariffId != "string" {
		query += fmt.Sprintf(" AND c.tariff_id = $%d", len(args)+1)
		args = append(args, req.TariffId)
	}
	if req.IsActive != "" && req.IsActive != "string" {
		if req.IsActive == "true" {
			query += `
				AND c.started_at <= NOW()
				AND (c.ended_at IS NULL OR c.ended_at > NOW())
			`
		} else if req.IsActive == "false" {
			query += `
				AND (
					c.started_at > NOW()
					OR (c.ended_at IS NOT NULL AND c.ended_at <= NOW())
				)
			`
		}
	}

	query += " ORDER BY c.created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting user courses: %w", err)
	}
	defer rows.Close()

	courses := m.UserCourseMobileList{}
	for rows.Next() {
		course := m.UserCourseMobileRes{}
		var totalCount int

		err := rows.Scan(
			&totalCount,
			&course.Id,
			&course.UserId,
			&course.UserName,
			&course.CourseId,
			&course.CourseName,
			&course.CourseImageUrl,
			&course.IsActive,
			&course.TariffId,
			&course.TariffName,
			&course.Duration,
			&course.StartedAt,
			&course.EndedAt,
			&course.CreatedAt,
			&course.UpdatedAt,
			&course.TotalLessons,
			&course.CompletedLessons,
			&course.Percentage,
		)

		if err != nil {
			return nil, nil, fmt.Errorf("error while scanning user course row: %w", err)
		}
		courses.UserCourses = append(courses.UserCourses, &course)
		courses.Count = totalCount
	}
	response := m.NewPaginationResponse(m.Pagination{Limit: courses.Count}, courses.Count, courses.UserCourses)
	return &response, &courses, nil
}
