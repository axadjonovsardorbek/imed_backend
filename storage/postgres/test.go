package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	m "imed/pkg/models"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAttemptNotFound = errors.New("attempt not found")
	ErrAttemptFinished = errors.New("attempt already finished")
	ErrTestTimeUp      = errors.New("test time is up")
	ErrInvalidOption   = errors.New("option does not belong to question")
)

// timeLimitGrace tarmoq kechikishi uchun qo'shimcha vaqt.
const timeLimitGrace = 10 * time.Second

type TestRepo struct {
	db *sql.DB
}

func NewTestRepo(db *sql.DB) *TestRepo {
	return &TestRepo{db: db}
}

func validateQuestion(q m.TestQuestionBody) error {
	if q.Text == "" {
		return fmt.Errorf("question text is empty")
	}
	if len(q.Options) < 2 {
		return fmt.Errorf("question %q must have at least 2 options", q.Text)
	}
	correct := 0
	for _, o := range q.Options {
		if o.Text == "" {
			return fmt.Errorf("question %q has empty option", q.Text)
		}
		if o.IsCorrect {
			correct++
		}
	}
	if correct != 1 {
		return fmt.Errorf("question %q must have exactly 1 correct option", q.Text)
	}
	return nil
}

func insertQuestion(ctx context.Context, tx *sql.Tx, testId string, q m.TestQuestionBody, orderNum int) (string, error) {
	if err := validateQuestion(q); err != nil {
		return "", err
	}
	if q.OrderNum > 0 {
		orderNum = q.OrderNum
	}

	id := uuid.NewString()
	_, err := tx.ExecContext(ctx, `
	INSERT INTO test_questions (id, test_id, text, image_url, order_num)
	VALUES ($1, $2, $3, NULLIF($4, ''), $5)
	`, id, testId, q.Text, q.ImageUrl, orderNum)
	if err != nil {
		return "", fmt.Errorf("error while creating question: %w", err)
	}

	if err := insertOptions(ctx, tx, id, q.Options); err != nil {
		return "", err
	}
	return id, nil
}

func insertOptions(ctx context.Context, tx *sql.Tx, questionId string, options []m.TestOptionBody) error {
	for i, o := range options {
		_, err := tx.ExecContext(ctx, `
		INSERT INTO test_options (question_id, text, is_correct, order_num)
		VALUES ($1, $2, $3, $4)
		`, questionId, o.Text, o.IsCorrect, i+1)
		if err != nil {
			return fmt.Errorf("error while creating option: %w", err)
		}
	}
	return nil
}

func nextQuestionOrder(ctx context.Context, tx *sql.Tx, testId string) (int, error) {
	var n int
	err := tx.QueryRowContext(ctx, `
	SELECT COALESCE(MAX(order_num), 0) FROM test_questions WHERE test_id = $1 AND deleted_at = 0
	`, testId).Scan(&n)
	return n, err
}

func (r *TestRepo) Create(ctx context.Context, req *m.TestCreateReq) (*m.ById, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	id := uuid.NewString()
	_, err = tx.ExecContext(ctx, `
	INSERT INTO tests (id, lesson_id, title, time_limit_sec, order_num)
	VALUES ($1, $2, $3, $4, $5)
	`, id, req.LessonId, req.Title, req.TimeLimitSec, req.OrderNum)
	if err != nil {
		return nil, fmt.Errorf("error while creating test: %w", err)
	}

	for i, q := range req.Questions {
		if _, err := insertQuestion(ctx, tx, id, q, i+1); err != nil {
			return nil, fmt.Errorf("question %d: %w", i+1, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &m.ById{Id: id}, nil
}

func (r *TestRepo) GetById(ctx context.Context, req *m.ById) (*m.TestRes, error) {
	var (
		res       m.TestRes
		titleJSON []byte
		qJSON     []byte
	)
	err := r.db.QueryRowContext(ctx, `
	SELECT
		t.id,
		t.lesson_id,
		t.title,
		t.time_limit_sec,
		t.order_num,
		t.created_at,
		t.updated_at,
		COALESCE((
			SELECT jsonb_agg(
				jsonb_build_object(
					'id', q.id,
					'text', q.text,
					'image_url', q.image_url,
					'order_num', q.order_num,
					'options', COALESCE((
						SELECT jsonb_agg(
							jsonb_build_object('id', o.id, 'text', o.text, 'is_correct', o.is_correct, 'order_num', o.order_num)
							ORDER BY o.order_num
						)
						FROM test_options o
						WHERE o.question_id = q.id
					), '[]'::jsonb)
				)
				ORDER BY q.order_num
			)
			FROM test_questions q
			WHERE q.test_id = t.id AND q.deleted_at = 0
		), '[]'::jsonb)
	FROM tests t
	WHERE t.id = $1 AND t.deleted_at = 0
	`, req.Id).Scan(
		&res.Id,
		&res.LessonId,
		&titleJSON,
		&res.TimeLimitSec,
		&res.OrderNum,
		&res.CreatedAt,
		&res.UpdatedAt,
		&qJSON,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("test not found with id: %s", req.Id)
	}
	if err != nil {
		return nil, fmt.Errorf("error while getting test: %w", err)
	}

	_ = json.Unmarshal(titleJSON, &res.Title)
	res.Questions = []m.TestQuestionRes{}
	_ = json.Unmarshal(qJSON, &res.Questions)
	res.QuestionCount = len(res.Questions)
	return &res, nil
}

func (r *TestRepo) GetAll(ctx context.Context, req *m.TestGetListReq) (*m.PaginationResponse, error) {
	query := `
	SELECT
		COUNT(t.id) OVER() AS total_count,
		t.id,
		t.lesson_id,
		t.title,
		t.time_limit_sec,
		t.order_num,
		(SELECT COUNT(*) FROM test_questions q WHERE q.test_id = t.id AND q.deleted_at = 0),
		t.created_at,
		t.updated_at
	FROM tests t
	WHERE t.deleted_at = 0
	`
	var args []interface{}
	if req.LessonId != "" {
		args = append(args, req.LessonId)
		query += fmt.Sprintf(" AND t.lesson_id = $%d", len(args))
	}
	args = append(args, req.Filter.Limit, req.Filter.Page)
	query += fmt.Sprintf(" ORDER BY t.order_num, t.created_at LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error while getting tests: %w", err)
	}
	defer rows.Close()

	list := m.TestList{Tests: []m.TestRes{}}
	for rows.Next() {
		var (
			t         m.TestRes
			titleJSON []byte
		)
		if err := rows.Scan(&list.Count, &t.Id, &t.LessonId, &titleJSON, &t.TimeLimitSec, &t.OrderNum, &t.QuestionCount, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error while scanning test: %w", err)
		}
		_ = json.Unmarshal(titleJSON, &t.Title)
		list.Tests = append(list.Tests, t)
	}

	res := m.NewPaginationResponse(req.Filter, list.Count, list.Tests)
	return &res, nil
}

func (r *TestRepo) Update(ctx context.Context, req *m.TestUpdateReq) (*m.Void, error) {
	res, err := r.db.ExecContext(ctx, `
	UPDATE tests
	SET title = $2, time_limit_sec = $3, order_num = $4, updated_at = now()
	WHERE id = $1 AND deleted_at = 0
	`, req.Id, req.Title, req.TimeLimitSec, req.OrderNum)
	if err != nil {
		return nil, fmt.Errorf("error while updating test: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, fmt.Errorf("test not found with id: %s", req.Id)
	}
	return &m.Void{}, nil
}

func (r *TestRepo) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	res, err := r.db.ExecContext(ctx, `
	UPDATE tests SET deleted_at = EXTRACT(EPOCH FROM now())::bigint WHERE id = $1 AND deleted_at = 0
	`, req.Id)
	if err != nil {
		return nil, fmt.Errorf("error while deleting test: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, fmt.Errorf("test not found with id: %s", req.Id)
	}
	return &m.Void{}, nil
}

// AddQuestions savollarni testga qo'shadi (bitta savol yoki Excel import).
// Hammasi bitta tranzaksiyada: xato bo'lsa hech narsa saqlanmaydi.
func (r *TestRepo) AddQuestions(ctx context.Context, testId string, questions []m.TestQuestionBody) ([]string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM tests WHERE id = $1 AND deleted_at = 0)`, testId).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("test not found with id: %s", testId)
	}

	order, err := nextQuestionOrder(ctx, tx, testId)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(questions))
	for i, q := range questions {
		order++
		id, err := insertQuestion(ctx, tx, testId, q, order)
		if err != nil {
			return nil, fmt.Errorf("question %d: %w", i+1, err)
		}
		ids = append(ids, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return ids, nil
}

// UpdateQuestion savol matnini yangilaydi va variantlarni to'liq almashtiradi.
// Eski variantlarga javob bergan urinishlar bo'lsa, variantlar o'chirilmaydi, savol yangidan yaratiladi.
func (r *TestRepo) UpdateQuestion(ctx context.Context, req *m.TestQuestionReq) (*m.Void, error) {
	if err := validateQuestion(req.TestQuestionBody); err != nil {
		return nil, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var (
		testId   string
		orderNum int
		answered bool
	)
	err = tx.QueryRowContext(ctx, `
	SELECT test_id, order_num, EXISTS (SELECT 1 FROM user_test_answers a WHERE a.question_id = q.id)
	FROM test_questions q
	WHERE id = $1 AND deleted_at = 0
	`, req.Id).Scan(&testId, &orderNum, &answered)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("question not found with id: %s", req.Id)
	}
	if err != nil {
		return nil, err
	}
	if req.OrderNum > 0 {
		orderNum = req.OrderNum
	}

	if answered {
		// Natijalar tarixi buzilmasligi uchun eski savol soft-delete qilinadi.
		if _, err := tx.ExecContext(ctx, `UPDATE test_questions SET deleted_at = EXTRACT(EPOCH FROM now())::bigint WHERE id = $1`, req.Id); err != nil {
			return nil, err
		}
		if _, err := insertQuestion(ctx, tx, testId, req.TestQuestionBody, orderNum); err != nil {
			return nil, err
		}
	} else {
		_, err = tx.ExecContext(ctx, `
		UPDATE test_questions SET text = $2, image_url = NULLIF($3, ''), order_num = $4, updated_at = now()
		WHERE id = $1
		`, req.Id, req.Text, req.ImageUrl, orderNum)
		if err != nil {
			return nil, fmt.Errorf("error while updating question: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM test_options WHERE question_id = $1`, req.Id); err != nil {
			return nil, err
		}
		if err := insertOptions(ctx, tx, req.Id, req.Options); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &m.Void{}, nil
}

func (r *TestRepo) DeleteQuestion(ctx context.Context, req *m.ById) (*m.Void, error) {
	res, err := r.db.ExecContext(ctx, `
	UPDATE test_questions SET deleted_at = EXTRACT(EPOCH FROM now())::bigint WHERE id = $1 AND deleted_at = 0
	`, req.Id)
	if err != nil {
		return nil, fmt.Errorf("error while deleting question: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, fmt.Errorf("question not found with id: %s", req.Id)
	}
	return &m.Void{}, nil
}

// ===== Mobile =====

func (r *TestRepo) GetByIdMobile(ctx context.Context, req *m.ByIdMobile) (*m.TestMobileRes, error) {
	var (
		res   m.TestMobileRes
		qJSON []byte
	)
	err := r.db.QueryRowContext(ctx, `
	SELECT
		t.id,
		t.lesson_id,
		COALESCE(t.title ->> $3, ''),
		t.time_limit_sec,
		(SELECT MAX(a.percent)::float8 FROM user_test_attempts a WHERE a.test_id = t.id AND a.user_id = $2 AND a.finished_at IS NOT NULL),
		COALESCE((
			SELECT jsonb_agg(
				jsonb_build_object(
					'id', q.id,
					'text', q.text,
					'image_url', q.image_url,
					'options', COALESCE((
						SELECT jsonb_agg(jsonb_build_object('id', o.id, 'text', o.text) ORDER BY o.order_num)
						FROM test_options o
						WHERE o.question_id = q.id
					), '[]'::jsonb)
				)
				ORDER BY q.order_num
			)
			FROM test_questions q
			WHERE q.test_id = t.id AND q.deleted_at = 0
		), '[]'::jsonb)
	FROM tests t
	WHERE t.id = $1 AND t.deleted_at = 0
	`, req.Id, req.UserId, req.Language).Scan(
		&res.Id,
		&res.LessonId,
		&res.Title,
		&res.TimeLimitSec,
		&res.BestPercent,
		&qJSON,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("test not found with id: %s", req.Id)
	}
	if err != nil {
		return nil, fmt.Errorf("error while getting test: %w", err)
	}

	res.Questions = []m.TestQuestionMobile{}
	_ = json.Unmarshal(qJSON, &res.Questions)
	return &res, nil
}

func (r *TestRepo) Start(ctx context.Context, req *m.ByIdMobile) (*m.TestStartRes, error) {
	var res m.TestStartRes
	err := r.db.QueryRowContext(ctx, `
	INSERT INTO user_test_attempts (user_id, test_id, total_count)
	SELECT $2, t.id, (SELECT COUNT(*) FROM test_questions q WHERE q.test_id = t.id AND q.deleted_at = 0)
	FROM tests t
	WHERE t.id = $1 AND t.deleted_at = 0
	RETURNING id, started_at, (SELECT time_limit_sec FROM tests WHERE id = $1)
	`, req.Id, req.UserId).Scan(&res.AttemptId, &res.StartedAt, &res.TimeLimitSec)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("test not found with id: %s", req.Id)
	}
	if err != nil {
		return nil, fmt.Errorf("error while starting test: %w", err)
	}
	return &res, nil
}

type attemptState struct {
	testId       string
	startedAt    time.Time
	finished     bool
	timeLimitSec *int
}

func (r *TestRepo) getAttempt(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}, attemptId, userId string) (*attemptState, error) {
	var (
		st         attemptState
		finishedAt sql.NullTime
	)
	err := q.QueryRowContext(ctx, `
	SELECT a.test_id, a.started_at, a.finished_at, t.time_limit_sec
	FROM user_test_attempts a
	JOIN tests t ON t.id = a.test_id
	WHERE a.id = $1 AND a.user_id = $2
	`, attemptId, userId).Scan(&st.testId, &st.startedAt, &finishedAt, &st.timeLimitSec)
	if err == sql.ErrNoRows {
		return nil, ErrAttemptNotFound
	}
	if err != nil {
		return nil, err
	}
	st.finished = finishedAt.Valid
	return &st, nil
}

func (s *attemptState) timeUp() bool {
	if s.timeLimitSec == nil || *s.timeLimitSec <= 0 {
		return false
	}
	deadline := s.startedAt.Add(time.Duration(*s.timeLimitSec)*time.Second + timeLimitGrace)
	return time.Now().After(deadline)
}

// Answer javobni saqlaydi va darhol to'g'ri/xato natijasini qaytaradi (T-09).
// Bir savolga faqat bitta javob: qayta yuborilsa birinchi javob qaytadi.
func (r *TestRepo) Answer(ctx context.Context, req *m.TestAnswerReq) (*m.TestAnswerRes, error) {
	st, err := r.getAttempt(ctx, r.db, req.AttemptId, req.UserId)
	if err != nil {
		return nil, err
	}
	if st.finished {
		return nil, ErrAttemptFinished
	}
	if st.timeUp() {
		return nil, ErrTestTimeUp
	}

	var (
		res       m.TestAnswerRes
		belongs   bool
		isCorrect bool
	)
	err = r.db.QueryRowContext(ctx, `
	SELECT
		EXISTS (
			SELECT 1 FROM test_options o
			JOIN test_questions q ON q.id = o.question_id
			WHERE o.id = $2 AND q.id = $1 AND q.test_id = $3
		),
		COALESCE((SELECT is_correct FROM test_options WHERE id = $2 AND question_id = $1), FALSE),
		COALESCE((SELECT id::text FROM test_options WHERE question_id = $1 AND is_correct LIMIT 1), '')
	`, req.QuestionId, req.OptionId, st.testId).Scan(&belongs, &isCorrect, &res.CorrectOptionId)
	if err != nil {
		return nil, fmt.Errorf("error while checking answer: %w", err)
	}
	if !belongs {
		return nil, ErrInvalidOption
	}

	err = r.db.QueryRowContext(ctx, `
	INSERT INTO user_test_answers (attempt_id, question_id, option_id, is_correct)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (attempt_id, question_id) DO UPDATE SET attempt_id = EXCLUDED.attempt_id
	RETURNING is_correct
	`, req.AttemptId, req.QuestionId, req.OptionId, isCorrect).Scan(&res.IsCorrect)
	if err != nil {
		return nil, fmt.Errorf("error while saving answer: %w", err)
	}
	return &res, nil
}

// Finish urinishni yakunlaydi: javob berilmagan savollar xato hisoblanadi (T-10).
func (r *TestRepo) Finish(ctx context.Context, req *m.TestAttemptReq) (*m.TestFinishRes, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	st, err := r.getAttempt(ctx, tx, req.AttemptId, req.UserId)
	if err != nil {
		return nil, err
	}

	res := m.TestFinishRes{AttemptId: req.AttemptId, TestId: st.testId}

	if !st.finished {
		_, err = tx.ExecContext(ctx, `
		UPDATE user_test_attempts a
		SET
			finished_at = now(),
			total_count = GREATEST(a.total_count, 1),
			correct_count = (SELECT COUNT(*) FROM user_test_answers ua WHERE ua.attempt_id = a.id AND ua.is_correct),
			percent = ROUND(
				(SELECT COUNT(*) FROM user_test_answers ua WHERE ua.attempt_id = a.id AND ua.is_correct) * 100.0
				/ GREATEST(a.total_count, 1), 2)
		WHERE a.id = $1
		`, req.AttemptId)
		if err != nil {
			return nil, fmt.Errorf("error while finishing test: %w", err)
		}
	}

	err = tx.QueryRowContext(ctx, `
	SELECT correct_count, total_count, percent::float8 FROM user_test_attempts WHERE id = $1
	`, req.AttemptId).Scan(&res.CorrectCount, &res.TotalCount, &res.Percent)
	if err != nil {
		return nil, err
	}

	// Keyingi mavzu: shu moduldagi keyingi dars, bo'lmasa keyingi modulning birinchi darsi.
	err = tx.QueryRowContext(ctx, `
	WITH cur AS (
		SELECT l.id, l.order_num, md.order_num AS module_order, md.course_id
		FROM tests t
		JOIN lessons l ON l.id = t.lesson_id
		JOIN modules md ON md.id = l.module_id
		WHERE t.id = $1
	)
	SELECT l.id::text
	FROM lessons l
	JOIN modules md ON md.id = l.module_id
	JOIN cur ON md.course_id = cur.course_id
	WHERE l.deleted_at = 0 AND md.deleted_at = 0
	  AND (md.order_num, l.order_num, l.id) > (cur.module_order, cur.order_num, cur.id)
	ORDER BY md.order_num, l.order_num, l.id
	LIMIT 1
	`, st.testId).Scan(&res.NextLessonId)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error while getting next lesson: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *TestRepo) GetByLessonMobile(ctx context.Context, lessonId, userId, lang string) ([]m.TestShortMobile, error) {
	rows, err := r.db.QueryContext(ctx, `
	SELECT
		t.id,
		COALESCE(t.title ->> $3, ''),
		t.time_limit_sec,
		(SELECT COUNT(*) FROM test_questions q WHERE q.test_id = t.id AND q.deleted_at = 0),
		(SELECT MAX(a.percent)::float8 FROM user_test_attempts a WHERE a.test_id = t.id AND a.user_id = $2 AND a.finished_at IS NOT NULL)
	FROM tests t
	WHERE t.lesson_id = $1 AND t.deleted_at = 0
	  AND EXISTS (SELECT 1 FROM test_questions q WHERE q.test_id = t.id AND q.deleted_at = 0)
	ORDER BY t.order_num, t.created_at
	`, lessonId, userId, lang)
	if err != nil {
		return nil, fmt.Errorf("error while getting lesson tests: %w", err)
	}
	defer rows.Close()

	list := []m.TestShortMobile{}
	for rows.Next() {
		var t m.TestShortMobile
		if err := rows.Scan(&t.Id, &t.Title, &t.TimeLimitSec, &t.QuestionCount, &t.BestPercent); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *TestRepo) History(ctx context.Context, req *m.TestHistoryReq) (*m.PaginationResponse, error) {
	rows, err := r.db.QueryContext(ctx, `
	SELECT
		COUNT(a.id) OVER(),
		a.id,
		a.test_id,
		COALESCE(t.title ->> $2, ''),
		t.lesson_id,
		a.correct_count,
		a.total_count,
		a.percent::float8,
		a.finished_at
	FROM user_test_attempts a
	JOIN tests t ON t.id = a.test_id
	WHERE a.user_id = $1 AND a.finished_at IS NOT NULL
	ORDER BY a.finished_at DESC
	LIMIT $3 OFFSET $4
	`, req.UserId, req.Language, req.Filter.Limit, req.Filter.Page)
	if err != nil {
		return nil, fmt.Errorf("error while getting test history: %w", err)
	}
	defer rows.Close()

	var (
		total int
		items = []m.TestHistoryItem{}
	)
	for rows.Next() {
		var it m.TestHistoryItem
		if err := rows.Scan(&total, &it.AttemptId, &it.TestId, &it.TestTitle, &it.LessonId, &it.CorrectCount, &it.TotalCount, &it.Percent, &it.FinishedAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}

	res := m.NewPaginationResponse(req.Filter, total, items)
	return &res, nil
}
