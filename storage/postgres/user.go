package postgres

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"imed/api/token"
	"imed/internal/helper"
	"imed/pkg/models"
	m "imed/pkg/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) Check(req *m.UserCheckReq) (bool, error) {
	var exists bool

	query := `
	SELECT EXISTS (
		SELECT 1
		FROM users
		WHERE deleted_at = 0
		AND phone_number = $1
	)
	`
	row := r.db.QueryRow(query, req.PhoneNumber)
	err := row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error while checking user existence: %w", err)
	}

	return true, nil
}

func (r *UserRepo) Login(req *m.UserLoginReq) (*m.TokenRes, error) {

	var id string
	var password string
	var role = "user"

	query := `
	SELECT
		id,
		password
	FROM
		users
	WHERE
		deleted_at = 0
	AND
		phone_number = $1
	`

	row := r.db.QueryRow(query, req.PhoneNumber)

	err := row.Scan(
		&id,
		&password,
	)

	if err == sql.ErrNoRows {

		if req.Password == "" {
			password := req.PhoneNumber
			req.Password = password
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}

		id := uuid.NewString()
		query := `
		INSERT INTO users (
			id,
			name,
			phone_number,
			password
		) VALUES ($1, $2, $3, $4)
		`

		_, err = r.db.Exec(query,
			id,
			req.Name,
			req.PhoneNumber,
			string(hashedPassword),
		)

		if err != nil {
			return nil, fmt.Errorf("error while registering user: %w", err)
		}

		tkn := token.GenerateJWTToken(id, role)

		return &m.TokenRes{
			AccessToken:  tkn.AccessToken,
			RefreshToken: tkn.RefreshToken,
			Role:         role,
			Id:           id,
		}, nil

	}

	if err != nil {
		return nil, fmt.Errorf("error while login: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	tkn := token.GenerateJWTToken(id, role)

	return &m.TokenRes{
		AccessToken:  tkn.AccessToken,
		RefreshToken: tkn.RefreshToken,
		Role:         role,
		Id:           id,
	}, nil
}

func (r *UserRepo) GetById(req *m.ById) (*m.UserRes, error) {
	query := `
	SELECT
		id,
		name,
		phone_number,
		fcm_token,
		image_url,
		language,
		created_at,
		updated_at
	FROM
		users
	WHERE
		deleted_at = 0
	AND
		id = $1
	`

	row := r.db.QueryRow(query, req.Id)

	var user m.UserRes
	err := row.Scan(
		&user.Id,
		&user.Name,
		&user.PhoneNumber,
		&user.FCMToken,
		&user.ImageUrl,
		&user.Language,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found with id: %s", req.Id)
	}

	if err != nil {
		return nil, fmt.Errorf("error while getting user by id: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) GetAll(req *m.UserGetListReq) (*m.PaginationResponse, *m.UserList, error) {
	query := `
	SELECT
		COUNT(id) OVER() AS total_count,
		id,
		name,
		phone_number,
		fcm_token,
		image_url,
		language,
		created_at,
		updated_at
	FROM
		users
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
		return nil, nil, fmt.Errorf("error while getting users: %w", err)
	}
	defer rows.Close()

	users := m.UserList{}

	for rows.Next() {
		user := m.UserRes{}
		var totalCount int
		err := rows.Scan(
			&totalCount,
			&user.Id,
			&user.Name,
			&user.PhoneNumber,
			&user.FCMToken,
			&user.ImageUrl,
			&user.Language,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error while scanning user row: %w", err)
		}
		users.Users = append(users.Users, user)
		users.Count = totalCount
	}

	response := m.NewPaginationResponse(req.Filter, users.Count, users.Users)

	return &response, &users, nil
}

func (r *UserRepo) Update(req *m.UserUpdateReq) (*m.Void, error) {
	query := `
	UPDATE users
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
	conditions = append(conditions, " fcm_token = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.FCMToken)

	conditions = append(conditions, " image_url = $"+strconv.Itoa(len(args)+1))
	args = append(args, req.ImageUrl)

	if req.Language != "" && req.Language != "string" {
		conditions = append(conditions, " language = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.Language)
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
		return nil, fmt.Errorf("error while updating user: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("user with id %s couldnt update", req.Id)
	}
	return &m.Void{}, nil
}

func (r *UserRepo) Delete(req *m.ById) (*m.Void, error) {
	query := `
	UPDATE users
	SET
		deleted_at = EXTRACT(EPOCH FROM NOW())
	WHERE
		id = $1
	AND
		deleted_at = 0
	`

	res, err := r.db.Exec(query, req.Id)
	if err != nil {
		return nil, fmt.Errorf("error while deleting user: %w", err)
	}

	if r, err := res.RowsAffected(); r == 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("user with id %s couldnt delete", req.Id)
	}

	return &m.Void{}, nil
}

func (r *UserRepo) ChangePassword(req *m.ChangePasswordReq) (*m.Void, error) {
	var curPass string

	queryCurrent := fmt.Sprintf("SELECT password FROM %s WHERE id = $1 AND deleted_at = 0", req.Role+"s")

	row := r.db.QueryRow(queryCurrent, req.Id)

	err := row.Scan(&curPass)

	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(curPass), []byte(req.OldPassword)); err != nil {
		return nil, fmt.Errorf("invalid current password: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error while hashing password: %w", err)
	}

	queryUpdate := fmt.Sprintf("UPDATE users %s SET password = $1 WHERE id = $2 AND deleted_at = 0", req.Role+"s")

	_, err = r.db.Exec(queryUpdate, string(hashedPassword), req.Id)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *UserRepo) RefreshPassword(req *m.RefreshPasswordReq) (*m.RefreshPassword, error) {

	password := helper.GeneratePassword()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error while hashing password: %w", err)
	}

	query := fmt.Sprintf(`
	UPDATE %s
	SET
		password = $1
	WHERE
		id = $2
	AND
		deleted_at = 0
	`, req.Role+"s")

	_, err = r.db.Exec(query, string(hashedPassword), req.UserId)

	if err != nil {
		return nil, err
	}

	return &m.RefreshPassword{
		UserId:   req.UserId,
		Password: password,
	}, nil
}

func (r *UserRepo) CreateActivity(req *m.UserActivityCreateReq) (*m.Void, error) {

	query := `
    INSERT INTO user_activities (user_id, day, activity)
    VALUES ($1, (now() AT TIME ZONE 'Asia/Tashkent')::date, $2)
    ON CONFLICT (user_id, day)
    DO UPDATE SET
		activity = user_activities.activity + EXCLUDED.activity,
        updated_at = now();
    `
	_, err := r.db.Exec(query, req.UserId, req.Activity)

	if err != nil {
		return nil, fmt.Errorf("error while creating/updating user activity: %w", err)
	}

	return nil, nil
}
func (r *UserRepo) GetActivities(req *m.UserActivityGetListReq) (*m.PaginationResponse, *m.UserActivityList, error) {

	if req.Type == "day"{
		query := `
		SELECT
			user_id,
			day,
			activity,
			updated_at
		FROM
			user_activities
		WHERE
			user_id = $1
		AND
			day = $2
		`

		row := r.db.QueryRow(query, req.UserId, req.Date)
		activities := m.UserActivityList{}
		activity := m.UserActivity{}

		err := row.Scan(
			&activity.UserId,
			&activity.Day,
			&activity.Activity,
			&activity.UpdatedAt,
		)
		
		if err == sql.ErrNoRows {
			activity.UserId = &req.UserId
			activity.Day = &req.Date
		}
		activities.Activities = append(activities.Activities, &activity)
		activities.Count = 1

		filter := models.Pagination{
			Limit: 1,
			Page:  0,
		}

		response := m.NewPaginationResponse(filter, activities.Count, activities.Activities)

		return &response, &activities, nil
	} else if req.Type == "range"{
		query := `
		SELECT
			user_id,
			day,
			activity,
			updated_at
		FROM
			user_activities
		WHERE
			user_id = $1
		AND
			day BETWEEN $2 AND $3
		ORDER BY day DESC
		`

		rows, err := r.db.Query(query, req.UserId, req.FromDate, req.ToDate)
		if err != nil {
			return nil, nil, fmt.Errorf("error while getting user activities: %w", err)
		}
		defer rows.Close()

		activities := m.UserActivityList{}

		for rows.Next() {
			activity := m.UserActivity{}
			err := rows.Scan(
				&activity.UserId,
				&activity.Day,
				&activity.Activity,
				&activity.UpdatedAt,
			)
			if err != nil {
				return nil, nil, fmt.Errorf("error while scanning user activity row: %w", err)
			}
			activities.Activities = append(activities.Activities, &activity)
		}
		activities.Count = len(activities.Activities)

		filter := models.Pagination{
			Limit: activities.Count,
			Page:  0,
		}

		response := m.NewPaginationResponse(filter, activities.Count, activities.Activities)

		return &response, &activities, nil

	}
	return nil, nil, nil
}
