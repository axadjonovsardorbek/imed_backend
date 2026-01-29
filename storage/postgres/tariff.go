package postgres

import (
	"database/sql"
	"fmt"
	m "imed/pkg/models"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type TariffRepo struct {
	db *sql.DB
}

func NewTariffRepo(db *sql.DB) *TariffRepo {
	return &TariffRepo{
		db: db,
	}
}

func (r *TariffRepo) Create(req *m.TariffCreateReq) (*m.ById, error) {
	id := uuid.NewString()
	query := `
	INSERT INTO tariffs (
		id,
		name,
		duration,
		description
	) VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(query,
		id,
		req.Name,
		req.Duration,
		req.Description,
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating tariff: %w", err)
	}

	return &m.ById{Id: id}, nil
}
func (r *TariffRepo) GetById(req *m.ById) (*m.TariffRes, error) {
	query := `
	SELECT 
		id,
		name,
		duration,
		description,
		created_at,
		updated_at
	FROM
		tariffs
	WHERE
		deleted_at = 0
	AND
		id = $1
	`

	row := r.db.QueryRow(query, req.Id)

	tariff := m.TariffRes{}
	err := row.Scan(
		&tariff.Id,
		&tariff.Name,
		&tariff.Duration,
		&tariff.Description,
		&tariff.CreatedAt,
		&tariff.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tariff not found with id: %s", req.Id)
	}

	if err != nil {
		return nil, fmt.Errorf("error while getting tariff by id: %w", err)
	}

	return &tariff, nil
}
func (r *TariffRepo) GetAll(req *m.TariffGetListReq) (*m.PaginationResponse, *m.TariffList, error) {
	query := `
	SELECT 
		COUNT(id) OVER() AS total_count,
		id,
		name,
		duration,
		description,
		created_at,
		updated_at
	FROM
		tariffs
	WHERE
		deleted_at = 0
	`

	var args []interface{}
	if req.Duration != 0 {
		query += fmt.Sprintf(" AND duration = $%d", len(args)+1)
		args = append(args, req.Duration)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting tariff: %w", err)
	}
	defer rows.Close()

	tariffs := m.TariffList{}

	for rows.Next() {

		tariff := m.TariffRes{}

		var totalCount int

		err := rows.Scan(
			&totalCount,
			&tariff.Id,
			&tariff.Name,
			&tariff.Duration,
			&tariff.Description,
			&tariff.CreatedAt,
			&tariff.UpdatedAt,
		)

		if err != nil {
			return nil, nil, fmt.Errorf("error while scanning tariff row: %w", err)
		}

		tariffs.Tariffs = append(tariffs.Tariffs, &tariff)
		tariffs.Count = totalCount
	}

	response := m.NewPaginationResponse(m.Pagination{Limit: tariffs.Count}, tariffs.Count, tariffs.Tariffs)

	return &response, &tariffs, nil
}
func (r *TariffRepo) Update(req *m.TariffUpdateReq) (*m.Void, error) {
	query := `
	UPDATE tariffs
	SET
	`

	var conditions []string
	var args []interface{}

	if req.Name != "" && req.Name != "string" {
		conditions = append(conditions, " name = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.Name)
	}

	conditions = append(conditions, " duration = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.Duration)

	conditions = append(conditions, " description = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.Description)

	if len(conditions) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	conditions = append(conditions, " updated_at = now()")
	query += strings.Join(conditions, ", ")
	query += " WHERE id = $" + strconv.Itoa(len(args)+1) + " AND deleted_at = 0"

	args = append(args, req.Id)

	res, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error while updating tariff: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("tariff with id %s couldnt update", req.Id)
	}
	return &m.Void{}, nil
}
func (r *TariffRepo) Delete(req *m.ById) (*m.Void, error) {
	query := `
	UPDATE tariffs
	SET
		deleted_at = EXTRACT(EPOCH FROM NOW())
	WHERE
		id = $1
	AND
		deleted_at = 0
	`

	res, err := r.db.Exec(query, req.Id)
	if err != nil {
		return nil, fmt.Errorf("error while deleting tariff: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("tariff with id %s couldnt delete", req.Id)
	}

	return &m.Void{}, nil
}
