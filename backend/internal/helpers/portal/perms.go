package portal

import (
	"context"
	"database/sql"
	"log"
)

func CanAccessCourse(db *sql.DB, studentID, courseID int64, ctx context.Context) bool {
	var belongs bool
	if err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1
		    FROM assigned_course_students
		    WHERE portal_user_id = $1
		    AND course_id = $2
		)
	`, studentID, courseID).Scan(&belongs); err != nil {
		log.Println(err)
		return false
	}

	return belongs
}
