package courses

import (
	"context"
	"database/sql"
	"strings"
)

func parseCoursePayload(name, description, color string) (string, string, string) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	color = strings.TrimSpace(strings.TrimPrefix(color, "#"))

	return name, description, color
}

func validateCoursePayload(name, description, color string) bool {
	// Validate
	if name == "" || len(name) > 32 {
		return false
	}
	if len(description) > 128 {
		return false
	}
	if color == "" || len(color) != 6 { // Has to be hex
		return false
	}

	return true
}

func checkNameConflict(name string, gradeID int64, tx *sql.Tx, ctx context.Context) (bool, error) {
	var conflict bool

	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM courses WHERE lower(name) = $1 AND grade_id = $2
		)
	`, strings.ToLower(name), gradeID).Scan(&conflict); err != nil {
		return false, err
	}

	return conflict, nil
}

func checkNameConflictExcept(name string, gradeID, courseID int64, tx *sql.Tx, ctx context.Context) (bool, error) {
	var conflict bool

	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM courses WHERE lower(name) = $1 AND grade_id = $2 AND id <> $3
		)
	`, strings.ToLower(name), gradeID, courseID).Scan(&conflict); err != nil {
		return false, err
	}

	return conflict, nil
}

func getCourseGradeAndSchoolID(db *sql.DB, ctx context.Context, courseID int64) (int64, int64, error) {
	// Get the course's school ID and grade ID
	var schoolID int64
	var gradeID int64
	if err := db.QueryRowContext(ctx, `
		SELECT s.id, c.grade_id
		FROM grades g
		JOIN courses c
		ON c.grade_id = g.id
		JOIN academic_years y
		ON g.academic_year_id = y.id
		JOIN schools s
		ON y.school_id = s.id
		WHERE c.id = $1
	`, courseID).Scan(&schoolID, &gradeID); err != nil {
		return 0, 0, err
	}

	return gradeID, schoolID, nil
}
