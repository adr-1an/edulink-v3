package submissions

import (
	"context"
	"database/sql"
)

func SubToSchoolID(subID int64, db *sql.DB, ctx context.Context) (int64, error) {
	var sID int64
	if err := db.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM assignment_submissions s
		JOIN assignments a ON s.assignment_id = a.id
		JOIN courses c ON a.course_id = c.id
		JOIN grades g ON c.grade_id = g.id
		JOIN academic_years y ON g.academic_year_id = y.id
		WHERE s.id = $1
	`, subID).Scan(&sID); err != nil {
		return 0, err
	}

	return sID, nil
}
