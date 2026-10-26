package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrNoAccess = errors.New("no access to this lesson")

// lessonAccessSQL $1 — dars, $2 — foydalanuvchi. Kurs sahifasidagi has_access bilan bir xil qoida:
// ochiq dars YOKI faol kurs obunasi YOKI faol premium YOKI ustoz.
const lessonAccessSQL = `
SELECT
	COALESCE(l.is_public, FALSE)
	OR EXISTS (
		SELECT 1 FROM user_courses uc
		WHERE uc.user_id = $2
		  AND uc.course_id = md.course_id
		  AND uc.deleted_at = 0
		  AND uc.started_at <= NOW()
		  AND (uc.ended_at IS NULL OR uc.ended_at > NOW())
	)
	OR EXISTS (
		SELECT 1 FROM user_subscriptions us
		WHERE us.user_id = $2 AND us.started_at <= NOW() AND us.ended_at > NOW()
	)
	OR EXISTS (SELECT 1 FROM users u WHERE u.id = $2 AND u.is_ustoz)
FROM lessons l
JOIN modules md ON md.id = l.module_id
WHERE l.id = $1 AND l.deleted_at = 0
`

// checkLessonAccess dars topilmasa yoki ruxsat bo'lmasa xato qaytaradi.
func checkLessonAccess(ctx context.Context, db *sql.DB, lessonId, userId string) error {
	var ok bool
	err := db.QueryRowContext(ctx, lessonAccessSQL, lessonId, userId).Scan(&ok)
	if err == sql.ErrNoRows {
		return fmt.Errorf("lesson not found with id: %s", lessonId)
	}
	if err != nil {
		return fmt.Errorf("error while checking lesson access: %w", err)
	}
	if !ok {
		return ErrNoAccess
	}
	return nil
}

func (r *LessonRepo) CheckAccess(ctx context.Context, lessonId, userId string) error {
	return checkLessonAccess(ctx, r.db, lessonId, userId)
}

func (r *TestRepo) CheckAccess(ctx context.Context, testId, userId string) error {
	var lessonId string
	err := r.db.QueryRowContext(ctx, `SELECT lesson_id FROM tests WHERE id = $1 AND deleted_at = 0`, testId).Scan(&lessonId)
	if err == sql.ErrNoRows {
		return fmt.Errorf("test not found with id: %s", testId)
	}
	if err != nil {
		return err
	}
	return checkLessonAccess(ctx, r.db, lessonId, userId)
}
