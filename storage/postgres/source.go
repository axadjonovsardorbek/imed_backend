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

type SourceRepo struct {
	db *sql.DB
}

func NewSourceRepo(db *sql.DB) *SourceRepo {
	return &SourceRepo{
		db: db,
	}
}

// CREATE TABLE IF NOT EXISTS sources (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     lesson_id UUID REFERENCES lessons(id),
//     name JSONB DEFAULT NULL,
//     url JSONB NOT NULL,
//     type VARCHAR(50) NOT NULL CHECK (type IN ('video', 'document', 'test')),
//     order_num BIGINT DEFAULT 1,
//     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
//     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
//     deleted_at BIGINT DEFAULT 0
// );

func (r *SourceRepo) Create(req *m.SourceCreateReq) (*m.ById, error) {
	id := uuid.NewString()
	query := `
	INSERT INTO sources (
		id,
		name,
		lesson_id,
		url,
		type,
		order_num
	) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(query,
		id,
		req.Name,
		req.LessonId,
		req.Url,
		req.Type,
		req.OrderNum,
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating source: %w", err)
	}

	return &m.ById{Id: id}, nil
}

func (r *SourceRepo) GetById(req *m.ById) (*m.SourceRes, error) {

	query := `
	SELECT 
		s.id,
		s.name,
		s.url,
		s.video_url,
		s.type,
		s.order_num,
		s.lesson_id,
		l.name,
		s.created_at,
		s.updated_at
	FROM
		sources s
	LEFT JOIN
		lessons l ON s.lesson_id = l.id
	WHERE
		s.deleted_at = 0
	AND
		s.id = $1
	`

	row := r.db.QueryRow(query, req.Id)

	source := m.SourceRes{}
	var nameJSON []byte
	var urlJSON []byte
	var videoUrlJSON []byte
	var lessonNameJSON []byte

	err := row.Scan(
		&source.Id,
		&nameJSON,
		&urlJSON,
		&videoUrlJSON,
		&source.Type,
		&source.OrderNum,
		&source.LessonId,
		&lessonNameJSON,
		&source.CreatedAt,
		&source.UpdatedAt,
	)

	json.Unmarshal(nameJSON, &source.Name)
	json.Unmarshal(urlJSON, &source.Url)
	json.Unmarshal(videoUrlJSON, &source.VideoUrl)
	json.Unmarshal(lessonNameJSON, &source.LessonName)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("source not found with id: %s", req.Id)
	}
	if err != nil {
		return nil, fmt.Errorf("error while getting source by id: %w", err)
	}
	return &source, nil
}

func (r *SourceRepo) GetAll(req *m.SourceGetListReq) (*m.PaginationResponse, *m.SourceList, error) {
	query := `
	SELECT 
		COUNT(s.id) OVER() AS total_count,
		s.id,
		s.name,
		s.url,
		s.video_url,
		s.type,
		s.order_num,
		s.lesson_id,
		l.name,
		s.created_at,
		s.updated_at
	FROM
		sources s
	LEFT JOIN
		lessons l ON s.lesson_id = l.id
	WHERE
		s.deleted_at = 0
	`

	var args []interface{}

	if req.Name != "" && req.Name != "string" {
		query += fmt.Sprintf(" AND s.name::text ILIKE $%d", len(args)+1)
		args = append(args, "%"+req.Name+"%")
	}
	if req.LessonId != "" && req.LessonId != "string" {
		query += fmt.Sprintf(" AND s.lesson_id = $%d", len(args)+1)
		args = append(args, req.LessonId)
	}
	if req.Type != "" && req.Type != "string" {
		query += fmt.Sprintf(" AND s.type = $%d", len(args)+1)
		args = append(args, req.Type)
	}

	query += " ORDER BY s.order_num ASC"
	args = append(args, req.Filter.Limit, req.Filter.Page)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting so: %w", err)
	}
	defer rows.Close()

	sources := m.SourceList{}
	for rows.Next() {
		source := m.SourceRes{}
		var nameJSON []byte
		var urlJSON []byte
		var videoUrlJSON []byte
		var lessonNameJSON []byte
		var totalCount int

		err := rows.Scan(
			&totalCount,
			&source.Id,
			&nameJSON,
			&urlJSON,
			&videoUrlJSON,
			&source.Type,
			&source.OrderNum,
			&source.LessonId,
			&lessonNameJSON,
			&source.CreatedAt,
			&source.UpdatedAt,
		)

		json.Unmarshal(nameJSON, &source.Name)
		json.Unmarshal(urlJSON, &source.Url)
		json.Unmarshal(videoUrlJSON, &source.VideoUrl)
		json.Unmarshal(lessonNameJSON, &source.LessonName)

		if err != nil {
			return nil, nil, fmt.Errorf("error while scanning source row: %w", err)
		}
		sources.Sources = append(sources.Sources, &source)
		sources.Count = totalCount
	}
	response := m.NewPaginationResponse(req.Filter, sources.Count, sources.Sources)
	return &response, &sources, nil
}

func (r *SourceRepo) Update(req *m.SourceUpdateReq) (*m.Void, error) {
	query := `
	UPDATE sources
	SET
	`

	var conditions []string
	var args []interface{}

	if req.Name != "" && req.Name != "string" {
		conditions = append(conditions, " name = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.Name)
	}

	conditions = append(conditions, " lesson_id = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.LessonId)

	conditions = append(conditions, " url = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.Url)

	conditions = append(conditions, " video_url = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.VideoUrl)

	conditions = append(conditions, " type = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.Type)

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
		return nil, fmt.Errorf("error while updating source: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("source with id %s couldnt update", req.Id)
	}
	return &m.Void{}, nil
}

func (r *SourceRepo) Delete(req *m.ById) (*m.Void, error) {
	query := `
	UPDATE sources
	SET
		deleted_at = EXTRACT(EPOCH FROM NOW())
	WHERE
		id = $1
	AND
		deleted_at = 0
	`

	res, err := r.db.Exec(query, req.Id)
	if err != nil {
		return nil, fmt.Errorf("error while deleting source: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("source with id %s couldnt delete", req.Id)
	}

	return &m.Void{}, nil
}

func (r *SourceRepo) UpdateVideoUrl(id string, videoUrl string) error {
	query := `
	UPDATE sources
	SET
		video_url = $1,
		updated_at = now()
	WHERE
		id = $2
	AND
		deleted_at = 0
	`

	_, err := r.db.Exec(query, videoUrl, id)
	if err != nil {
		return fmt.Errorf("error while updating video_url: %w", err)
	}

	return nil
}

// func (r *SourceRepo)GetByIdMobile(req *m.ById) (*m.SourceMobileRes, error){
// 	return nil, nil
// }
// func (r *SourceRepo)GetAllMobile(req *m.SourceGetListMobileReq) (*m.PaginationResponse, *m.SourceMobileList, error){
// 	return nil, nil, nil
// }
