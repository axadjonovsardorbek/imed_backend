package postgres

import (
	"database/sql"
	"fmt"

	"imed/api/token"
	m "imed/pkg/models"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AdminRepo struct {
	db  *sql.DB
	rdb *redis.Client
}

func NewAdminRepo(db *sql.DB, rdb *redis.Client) *AdminRepo {
	return &AdminRepo{
		db:  db,
		rdb: rdb,
	}
}

func (r *AdminRepo) Login(req *m.LoginReq) (*m.TokenRes, error) {

	var id string
	var password string

	query := `
	SELECT 
		id,
		password
	FROM
		admins
	WHERE
		login = $1
	`

	row := r.db.QueryRow(query, req.Login)

	err := row.Scan(
		&id,
		&password,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid phone: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("error while login: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	tkn, err := token.GenerateJWTToken(id, "admin", r.rdb)
	if err != nil {
		return nil, err
	}

	return &m.TokenRes{
		AccessToken:  tkn.AccessToken,
		RefreshToken: tkn.RefreshToken,
		Role:         "admin",
		Id:           id,
	}, nil
}

func (r *AdminRepo) GetById(req *m.ById) (*m.AdminRes, error) {
	query := `
	SELECT
		a.id,
		a.login,
		a.created_at,
		a.updated_at
	FROM
		admins a
	WHERE
		a.id = $1
	`
	admin := m.AdminRes{}

	row := r.db.QueryRow(query, req.Id).Scan(
		&admin.Id,
		&admin.Login,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)

	if row == sql.ErrNoRows {
		return nil, nil
	}
	if row != nil {
		return nil, fmt.Errorf("error while getting admin by id: %w", row)
	}

	return &admin, nil
}
