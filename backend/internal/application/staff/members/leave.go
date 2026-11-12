package members

import (
	"context"
	"fmt"
	"log"

	"app/internal/helpers/staff/schools"
)

type LeaveSchoolStaffInput struct {
	UserID   int64
	SchoolID int64
}

type LeaveSchoolStaffResult struct {
	// NoRowsAffected preserves the result of the original write without changing transaction flow.
	NoRowsAffected bool
}

func (s *Service) LeaveSchoolStaff(ctx context.Context, i *LeaveSchoolStaffInput) (*LeaveSchoolStaffResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	userID := i.UserID
	schoolID := i.SchoolID
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var staffRowExists bool
	var userName string
	var userEmail string
	if err := tx.QueryRowContext(ctx, `
		WITH staff_exists AS (
		    SELECT EXISTS (
				SELECT 1 FROM school_staff ss
				JOIN schools s ON ss.school_id = s.id
				WHERE ss.school_id = $2
				AND ss.user_id = $1
				AND s.owner_id <> $1
			) AS exists
		)
		SELECT staff_exists.exists, users.name, users.email
		FROM staff_exists
		JOIN users ON users.id = $1
		`, userID, schoolID).Scan(&staffRowExists, &userName, &userEmail); err != nil {
		log.Println(err)
		return nil, err
	}

	if !staffRowExists {
		return nil, ErrForbidden
	}

	res, err := tx.ExecContext(ctx, `
		DELETE FROM school_staff WHERE user_id = $1 AND school_id = $2
	`, userID, schoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	details := fmt.Sprintf("%s <%s> left the school's staff members.", userName, userEmail)
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionSchoolLeave, schools.TypeDelete, "User left", "{user} left the school.", tx, ctx, s.Sf, details); err != nil {
		log.Println(err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	affectedRows, err = res.RowsAffected()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	noRowsAffected = noRowsAffected || affectedRows == 0

	return &LeaveSchoolStaffResult{NoRowsAffected: noRowsAffected}, nil
}
