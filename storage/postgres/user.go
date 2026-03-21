package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"imed/api/token"
	"imed/internal/helper"
	"imed/pkg/models"
	m "imed/pkg/models"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo struct {
	db  *sql.DB
	rdb *redis.Client
}

func NewUserRepo(db *sql.DB, rdb *redis.Client) *UserRepo {
	return &UserRepo{
		db:  db,
		rdb: rdb,
	}
}

func (r *UserRepo) GetTokens(ctx context.Context, f m.TokenFilter) ([]string, error) {
	q := `SELECT DISTINCT u.fcm_token FROM users u `
	args := make([]any, 0, 2)

	if f.CourseID != "" {
		q += `JOIN user_courses uc ON uc.user_id = u.id `
	}

	q += `
WHERE u.deleted_at = 0
  AND u.fcm_token IS NOT NULL
  AND u.fcm_token <> ''
`

	if f.CourseID != "" {
		args = append(args, f.CourseID)
		q += fmt.Sprintf(`
  AND uc.deleted_at = 0
  AND uc.course_id = $%d
`, len(args))
	}

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("get fcm tokens query: %w", err)
	}
	defer rows.Close()

	tokens := make([]string, 0, 256)
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, fmt.Errorf("scan fcm token: %w", err)
		}
		tokens = append(tokens, token)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate fcm tokens: %w", err)
	}

	fmt.Println("Resolved FCM tokens:", len(tokens))
	slog.Info("Resolved FCM tokens", "count", len(tokens))

	return tokens, nil
}

func (r *UserRepo) GetToken(user_id string) (*string, error) {
	query := `
	SELECT fcm_token
	FROM users
	WHERE deleted_at = 0
	AND id = $1
	AND fcm_token IS NOT NULL
	AND fcm_token <> ''
	`

	var token *string
	err := r.db.QueryRow(query, user_id).Scan(&token)
	if err != nil {
		return nil, fmt.Errorf("error while getting fcm token: %w", err)
	}

	return token, nil
}
func (r *UserRepo) CreateUser(req *m.UserLoginReq) (*m.ById, error) {
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

	return &m.ById{Id: id}, nil
}

func (r *UserRepo) Check(req *m.UserCheckReq) (*m.UserCheckRes, error) {
	var res m.UserCheckRes

	query := `
	SELECT EXISTS (
		SELECT 1
		FROM users
		WHERE deleted_at = 0
		AND (phone_number = $1 OR email = $1)
	)
	`
	row := r.db.QueryRow(query, req.Identifier)
	err := row.Scan(&res.HasAccount)
	if err != nil {
		return &res, fmt.Errorf("error while checking user existence: %w", err)
	}

	return &res, nil
}

func (r *UserRepo) EmailLogin(req *m.ConfirmationReq) (*m.TokenRes, error) {

	var count int
	var user_id string
	var tkn *token.Tokens

	req.Identifier = strings.ToLower(req.Identifier)

	query := `
	SELECT 
		COUNT(id) OVER() as count,
		id
	FROM
		users
	WHERE
		email = $1
	AND
		deleted_at = 0
	GROUP BY id
	`

	row := r.db.QueryRow(query, req.Identifier)

	err := row.Scan(
		&count,
		&user_id,
	)

	if err != nil {
		fmt.Println("error while checking user existence:", err)
	}

	if count == 1 && user_id != "string" {
		tkn = token.GenerateJWTToken(user_id, "user", r.rdb)
		return &m.TokenRes{
			AccessToken:  tkn.AccessToken,
			RefreshToken: tkn.RefreshToken,
			Role:         "user",
			Id:           user_id,
		}, nil
	} else {
		id := uuid.New().String()

		query := `
			INSERT INTO users (
				id,
				email,
				name
			) VALUES($1, $2, $3)
			`

		_, err = r.db.Exec(query, id, req.Identifier, "Foydalanuvchi")

		if err != nil {
			return nil, err
		}
		tkn = token.GenerateJWTToken(id, "user", r.rdb)
		return &m.TokenRes{
			AccessToken:  tkn.AccessToken,
			RefreshToken: tkn.RefreshToken,
			Role:         "user",
			Id:           id,
		}, nil
	}
}

func (r *UserRepo) TelegramLogin(req *m.TelegramLoginReq) (*m.TokenRes, error) {

	var count int
	var user_id string
	var tkn *token.Tokens

	query := `
	SELECT 
		COUNT(id) OVER() as count,
		id
	FROM
		users
	WHERE
		phone_number = $1
	AND
		deleted_at = 0
	GROUP BY id
	`

	row := r.db.QueryRow(query, req.Phone)

	err := row.Scan(
		&count,
		&user_id,
	)

	if err != nil {
		fmt.Println("error while checking user existence:", err)
	}

	if count == 1 && user_id != "string" {
		tkn = token.GenerateJWTToken(user_id, "user", r.rdb)
		return &m.TokenRes{
			AccessToken:  tkn.AccessToken,
			RefreshToken: tkn.RefreshToken,
			Role:         "user",
			Id:           user_id,
		}, nil
	} else {
		id := uuid.New().String()

		query := `
			INSERT INTO users (
				id,
				phone_number,
				name
			) VALUES($1, $2, $3)
			`

		_, err = r.db.Exec(query, id, req.Phone, req.Name)

		if err != nil {
			return nil, err
		}
		tkn = token.GenerateJWTToken(id, "user", r.rdb)
		return &m.TokenRes{
			AccessToken:  tkn.AccessToken,
			RefreshToken: tkn.RefreshToken,
			Role:         "user",
			Id:           id,
		}, nil
	}
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

		tkn := token.GenerateJWTToken(id, role, r.rdb)

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

	var valid bool

	if strings.HasPrefix(password, "$2") {
		valid = bcrypt.CompareHashAndPassword(
			[]byte(password),
			[]byte(req.Password),
		) == nil

	} else if strings.HasPrefix(password, "pbkdf2_sha256$") {
		valid, _ = helper.CheckDjangoPBKDF2(req.Password, password)
	}

	if !valid {
		return nil, fmt.Errorf("invalid credentials")
	}

	tkn := token.GenerateJWTToken(id, role, r.rdb)

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
		email,
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
		&user.Email,
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
		email,
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
			&user.Email,
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
		var exists bool
		checkQuery := `
			SELECT EXISTS(
				SELECT 1
				FROM users
				WHERE phone_number = $1
				  AND id <> $2
				  AND deleted_at = 0
			)
		`
		err := r.db.QueryRow(checkQuery, req.PhoneNumber, req.Id).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("error while checking phone number: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("bu telefon raqam allaqachon mavjud")
		}

		conditions = append(conditions, "phone_number = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.PhoneNumber)
	}
	if req.FCMToken != "" && req.FCMToken != "string" {
		conditions = append(conditions, " fcm_token = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.FCMToken)
	}
	if req.ImageUrl != "" && req.ImageUrl != "string" {
		conditions = append(conditions, " image_url = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.ImageUrl)
	}
	if req.Language != "" && req.Language != "string" {
		conditions = append(conditions, " language = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.Language)
	}
	if req.Email != "" && req.Email != "string" {
		var exists bool
		checkQuery := `
			SELECT EXISTS(
				SELECT 1
				FROM users
				WHERE email = $1
				  AND id <> $2
				  AND deleted_at = 0
			)
		`
		err := r.db.QueryRow(checkQuery, req.Email, req.Id).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("error while checking email: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("bu email allaqachon mavjud")
		}

		conditions = append(conditions, "email = $"+strconv.Itoa(len(args)+1))
		args = append(args, req.Email)
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

	var queryCurrent string

	if req.Role == "user" {
		queryCurrent = "SELECT password FROM users WHERE id = $1 AND deleted_at = 0"
	} else if req.Role == "admin" {
		queryCurrent = "SELECT password FROM admins WHERE id = $1"
	} else {
		return nil, fmt.Errorf("incorrect role")
	}

	row := r.db.QueryRow(queryCurrent, req.Id)

	err := row.Scan(&curPass)

	if err != nil {
		return nil, err
	}

	var valid bool

	if strings.HasPrefix(curPass, "$2") {
		valid = bcrypt.CompareHashAndPassword(
			[]byte(curPass),
			[]byte(req.OldPassword),
		) == nil

	} else if strings.HasPrefix(curPass, "pbkdf2_sha256$") {
		valid, _ = helper.CheckDjangoPBKDF2(req.OldPassword, curPass)
	}

	if !valid {
		return nil, fmt.Errorf("invalid credentials")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error while hashing password: %w", err)
	}

	var queryUpdate string
	if req.Role == "user" {
		queryUpdate = "UPDATE users SET password = $1 WHERE id = $2 AND deleted_at = 0"
	} else if req.Role == "admin" {
		queryUpdate = "UPDATE admins SET password = $1 WHERE id = $2"
	} else {
		return nil, fmt.Errorf("incorrect role")
	}

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

	var query string

	if req.Role == "user" {
		query = "UPDATE users SET password = $1 WHERE id = $2 AND deleted_at = 0"
	} else if req.Role == "admin" {
		query = "UPDATE admins SET password = $1 WHERE id = $2"
	} else if req.Role == "teacher" {
		query = "UPDATE teachers SET password = $1 WHERE id = $2 AND deleted_at = 0"
	} else {
		return nil, fmt.Errorf("incorrect role")
	}

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
func (r *UserRepo) GetActivities(req *m.UserActivityGetListReq) (*m.ActivityStatsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}
	if req.UserId == "" {
		return nil, fmt.Errorf("user_id required")
	}
	if req.Type == "" {
		return nil, fmt.Errorf("type required")
	}

	res := &models.ActivityStatsResponse{
		UserID: req.UserId,
		Type:   req.Type,
		Items:  make([]models.ActivityStatItem, 0, 32),
	}

	switch req.Type {

	case "day":
		query := `
	SELECT
		COALESCE(NULLIF($2, '')::date, CURRENT_DATE) AS day,
		COALESCE(ua.activity, 0) AS value
	FROM (SELECT 1) x
	LEFT JOIN user_activities ua
	ON ua.user_id = $1
	AND ua.day = COALESCE(NULLIF($2, '')::date, CURRENT_DATE)
	
		`

		var day time.Time
		var value int64

		if err := r.db.QueryRow(query, req.UserId, req.Date).Scan(&day, &value); err != nil {
			return nil, fmt.Errorf("day query error: %w", err)
		}

		res.Items = append(res.Items, models.ActivityStatItem{
			Date:  day.Format("2006-01-02"),
			Value: value,
		})
		res.Total = value
		return res, nil

	case "week", "month", "range":
		var start interface{}
		var end interface{}

		if req.Type == "range" {
			if req.FromDate == "" || req.ToDate == "" {
				return nil, fmt.Errorf("from_date and to_date are required for range")
			}
			start = req.FromDate
			end = req.ToDate
		} else if req.Type == "week" {
			queryWeek := `
			WITH r AS (
			  SELECT $1::uuid AS user_id,
			         date_trunc('week', CURRENT_DATE)::date AS start_day,
			         CURRENT_DATE AS end_day
			),
			series AS (
			  SELECT
				d.day::date AS day,
				COALESCE(ua.activity, 0) AS value
			  FROM r
			  JOIN LATERAL generate_series(r.start_day, r.end_day, interval '1 day') AS d(day) ON TRUE
			  LEFT JOIN user_activities ua
				ON ua.user_id = r.user_id
			   AND ua.day = d.day::date
			)
			SELECT day, value, SUM(value) OVER () AS total
			FROM series
			ORDER BY day DESC
			`

			rows, err := r.db.Query(queryWeek, req.UserId)
			if err != nil {
				return nil, fmt.Errorf("week query error: %w", err)
			}
			defer rows.Close()

			items, total, err := scanKeyValueTotalRows(rows)
			if err != nil {
				return nil, fmt.Errorf("week scan error: %w", err)
			}
			res.Items = items
			res.Total = total
			return res, nil

		} else {
			queryMonth := `
			WITH r AS (
			  SELECT $1::uuid AS user_id,
			         date_trunc('month', CURRENT_DATE)::date AS start_day,
			         CURRENT_DATE AS end_day
			),
			series AS (
			  SELECT
				d.day::date AS day,
				COALESCE(ua.activity, 0) AS value
			  FROM r
			  JOIN LATERAL generate_series(r.start_day, r.end_day, interval '1 day') AS d(day) ON TRUE
			  LEFT JOIN user_activities ua
				ON ua.user_id = r.user_id
			   AND ua.day = d.day::date
			)
			SELECT day, value, SUM(value) OVER () AS total
			FROM series
			ORDER BY day DESC
			`

			rows, err := r.db.Query(queryMonth, req.UserId)
			if err != nil {
				return nil, fmt.Errorf("month query error: %w", err)
			}
			defer rows.Close()

			items, total, err := scanKeyValueTotalRows(rows)
			if err != nil {
				return nil, fmt.Errorf("month scan error: %w", err)
			}
			res.Items = items
			res.Total = total
			return res, nil
		}

		queryRange := `
		WITH r AS (
		  SELECT
			$1::uuid AS user_id,
			$2::date AS start_day,
			LEAST($3::date, CURRENT_DATE) AS end_day
		),
		series AS (
		  SELECT
			d.day::date AS day,
			COALESCE(ua.activity, 0) AS value
		  FROM r
		  JOIN LATERAL generate_series(r.start_day, r.end_day, interval '1 day') AS d(day)
			ON r.start_day <= r.end_day
		  LEFT JOIN user_activities ua
			ON ua.user_id = r.user_id
		   AND ua.day = d.day::date
		)
		SELECT day, value, SUM(value) OVER () AS total
		FROM series
		ORDER BY day DESC
		`

		rows, err := r.db.Query(queryRange, req.UserId, start, end)
		if err != nil {
			return nil, fmt.Errorf("range query error: %w", err)
		}
		defer rows.Close()

		items, total, err := scanKeyValueTotalRows(rows)
		if err != nil {
			return nil, fmt.Errorf("range scan error: %w", err)
		}
		res.Items = items
		res.Total = total
		return res, nil

	case "year":
		query := `
		WITH months AS (
		  SELECT generate_series(
				   date_trunc('year', CURRENT_DATE)::date,
				   date_trunc('month', CURRENT_DATE)::date,
				   interval '1 month'
				 )::date AS month_start
		),
		agg AS (
		  SELECT
			m.month_start AS day,
			COALESCE(SUM(ua.activity), 0) AS value
		  FROM months m
		  LEFT JOIN user_activities ua
			ON ua.user_id = $1
		   AND ua.day >= m.month_start
		   AND ua.day <  (m.month_start + interval '1 month')::date
		   AND ua.day <= CURRENT_DATE
		  GROUP BY m.month_start
		)
		SELECT day, value, SUM(value) OVER () AS total
		FROM agg
		ORDER BY day DESC
		`

		rows, err := r.db.Query(query, req.UserId)
		if err != nil {
			return nil, fmt.Errorf("year query error: %w", err)
		}
		defer rows.Close()

		items, total, err := scanKeyValueTotalRows(rows)
		if err != nil {
			return nil, fmt.Errorf("year scan error: %w", err)
		}
		res.Items = items
		res.Total = total
		return res, nil

	default:
		return nil, fmt.Errorf("unsupported type: %s", req.Type)
	}
}

func (r *UserRepo) GetRating(req *m.RatingRequest) (*m.RatingResponse, error) {

	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 10
	}
	if req.Type == "" {
		req.Type = "total"
	}

	loc := time.FixedZone("UTC+5", 5*3600)
	rt := m.RatingType(req.Type)
	start, end := calcRange(rt, loc)

	res := &m.RatingResponse{
		Type:  rt,
		Limit: req.Limit,
		Items: make([]m.RatingUserItem, 0, req.Limit),
	}

	query := `
	WITH scores AS (
		SELECT
		  u.id AS user_id,
		  u.name AS name,
		  u.image_url AS avatar,
		  COALESCE(SUM(ua.activity), 0) AS activity
		FROM users u
		LEFT JOIN user_activities ua
		  ON ua.user_id = u.id
		 AND ua.day >= $1::date
		 AND ua.day <= $2::date
		GROUP BY u.id, u.name, u.image_url
	),
	ranked AS (
		SELECT
		  user_id,
		  name,
		  avatar,
		  activity,
		  ROW_NUMBER() OVER (
			ORDER BY activity DESC, name ASC, user_id ASC
		  ) AS rank
		FROM scores
	),
	topn AS (
		SELECT user_id, name, avatar, activity, rank
		FROM ranked
		ORDER BY rank
		LIMIT $3
	),
	me AS (
		SELECT user_id, name, avatar, activity, rank
		FROM ranked
		WHERE user_id = $4::uuid
	)
	SELECT 'top' AS kind, user_id, name, avatar, activity, rank FROM topn
	UNION ALL
	SELECT 'me'  AS kind, user_id, name, avatar, activity, rank FROM me;
	`

	rows, err := r.db.Query(query, start, end, req.Limit, req.UserId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var kind string
		var it m.RatingUserItem

		if err := rows.Scan(&kind, &it.UserID, &it.Name, &it.ImageUrl, &it.Activity, &it.Rank); err != nil {
			return nil, err
		}

		if kind == "me" {
			res.Me = &it
		} else {
			res.Items = append(res.Items, it)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *UserRepo) GetTop(req *m.RatingRequest) ([]m.RatingUserItem, error) {
	qTopRating := `
	WITH scores AS (
	  SELECT
	    u.id AS user_id,
	    u.name AS name,
	    u.image_url AS avatar,
	    COALESCE(SUM(ua.activity), 0) AS activity
	  FROM users u
	  LEFT JOIN user_activities ua
	    ON ua.user_id = u.id
	   AND ua.day >= $1::date
	   AND ua.day <= $2::date
	  GROUP BY u.id, u.name, u.image_url
	),
	ranked AS (
	  SELECT
	    user_id, name, avatar, activity,
	    ROW_NUMBER() OVER (ORDER BY activity DESC, name ASC, user_id ASC) AS rank
	  FROM scores
	)
	SELECT user_id, name, avatar, activity, rank, (user_id = $3) as is_me
	FROM ranked
	ORDER BY rank
	LIMIT $4
	`

	loc := time.FixedZone("UTC+5", 5*3600)
	rt := m.RatingType(req.Type)
	start, end := calcRange(rt, loc)

	rows, err := r.db.Query(qTopRating, start, end, req.UserId, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("top rating query: %w", err)
	}
	defer rows.Close()

	items := make([]m.RatingUserItem, 0, req.Limit)
	for rows.Next() {
		var it m.RatingUserItem

		if err := rows.Scan(&it.UserID, &it.Name, &it.ImageUrl, &it.Activity, &it.Rank, &it.IsMe); err != nil {
			return nil, fmt.Errorf("top rating scan: %w", err)
		}

		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("top rating rows: %w", err)
	}

	slog.Debug("GetTop rating", "type", req.Type, "limit", req.Limit, "items_count", len(items))
	fmt.Println("GetTop rating", "type", req.Type, "limit", req.Limit, "items_count", len(items))
	return items, nil
}

func (r *UserRepo) GetMe(req *m.RatingRequest) (*m.RatingUserItem, error) {
	qMeRating := `
	WITH scores AS (
	  SELECT
	    u.id AS user_id,
	    u.name AS name,
	    u.image_url AS avatar,
	    COALESCE(SUM(ua.activity), 0) AS activity
	  FROM users u
	  LEFT JOIN user_activities ua
	    ON ua.user_id = u.id
	   AND ua.day >= $1::date
	   AND ua.day <= $2::date
	  GROUP BY u.id, u.name, u.image_url
	),
	ranked AS (
	  SELECT
	    user_id, name, avatar, activity,
	    ROW_NUMBER() OVER (ORDER BY activity DESC, name ASC, user_id ASC) AS rank
	  FROM scores
	)
	SELECT user_id, name, avatar, activity, rank
	FROM ranked
	WHERE user_id = $3::uuid
	`

	var it models.RatingUserItem

	loc := time.FixedZone("UTC+5", 5*3600)
	rt := m.RatingType(req.Type)
	start, end := calcRange(rt, loc)

	err := r.db.QueryRow(qMeRating, start, end, req.UserId).
		Scan(&it.UserID, &it.Name, &it.ImageUrl, &it.Activity, &it.Rank)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("me rating query: %w", err)
	}

	it.IsMe = true

	slog.Debug("GetMe rating", "type", req.Type, "user_id", req.UserId, "activity", it.Activity, "rank", it.Rank)
	fmt.Println("GetMe rating", "type", req.Type, "user_id", req.UserId, "activity", it.Activity, "rank", it.Rank)
	return &it, nil
}

func scanKeyValueTotalRows(rows *sql.Rows) (items []models.ActivityStatItem, total int64, err error) {
	for rows.Next() {
		var day time.Time
		var value int64
		var rowTotal sql.NullInt64

		if err := rows.Scan(&day, &value, &rowTotal); err != nil {
			return nil, 0, err
		}
		if rowTotal.Valid {
			total = rowTotal.Int64
		}
		items = append(items, models.ActivityStatItem{
			Date:  day.Format("2006-01-02"),
			Value: value,
		})
	}
	return items, total, rows.Err()
}

func calcRange(t m.RatingType, loc *time.Location) (time.Time, time.Time) {
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	switch t {
	case m.RatingDay:
		return today, today

	case m.RatingWeek:
		wd := int(today.Weekday())
		if wd == 0 {
			wd = 7
		}
		start := today.AddDate(0, 0, -(wd - 1))
		return start, today

	case m.RatingMonth:
		start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, loc)
		return start, today

	case m.RatingYear:
		start := time.Date(today.Year(), time.January, 1, 0, 0, 0, 0, loc)
		return start, today

	default:
		start := time.Date(1970, 1, 1, 0, 0, 0, 0, loc)
		return start, today
	}
}
