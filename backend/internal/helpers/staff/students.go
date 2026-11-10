package staffhelpers

import (
	"app/internal/helpers"
	"context"
	"database/sql"
)

func StudentToSchoolID(ctx context.Context, db *sql.DB, studentID int64) (int64, error) {
	var sID int64
	if err := db.QueryRowContext(ctx, `
		SELECT school_id FROM portal_users WHERE id = $1 AND account_type = $2
	`, studentID, helpers.AccTypeStudent).Scan(&sID); err != nil {
		return 0, err
	}

	return sID, nil
}
