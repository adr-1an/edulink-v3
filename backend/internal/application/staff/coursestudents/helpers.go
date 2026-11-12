package coursestudents

import (
	"database/sql"
)

func getCourseSchoolID(courseID int64, tx *sql.Tx) (int64, error) {
	var schoolID int64
	if err := tx.QueryRow(`
		SELECT y.school_id
		FROM courses c
		JOIN grades g
		ON c.grade_id = g.id
		JOIN academic_years y
		ON g.academic_year_id = y.id
		WHERE c.id = $1
	`, courseID).Scan(&schoolID); err != nil {
		return 0, err
	}

	return schoolID, nil
}
