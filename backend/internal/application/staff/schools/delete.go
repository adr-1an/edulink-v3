package schools

import (
	"app/internal/helpers/staff/schools"
	"context"
	"log"
	"strconv"
)

type DeleteSchoolInput struct {
	UserID   int64
	SchoolID int64
}

func (s *Service) DeleteSchool(ctx context.Context, i *DeleteSchoolInput) error {
	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var schoolName string
	if err := tx.QueryRowContext(ctx, `
		SELECT name FROM schools WHERE id = $1
	`, i.SchoolID).Scan(&schoolName); err != nil {
		log.Println(err)
		return err
	}

	// Delete school (update deleted_at)
	if _, err := tx.ExecContext(ctx, `
		UPDATE schools
		SET deleted_at = NOW()
		WHERE id = $1
	`, i.SchoolID); err != nil {
		log.Println(err)
		return err
	}

	// Log action
	if err := schools.StoreSchoolLog(i.SchoolID, i.UserID, schools.ActionSchoolDelete, schools.TypeDelete, "School deleted", "{user} deleted the school '"+schoolName+"'.", tx, ctx, s.Sf, "{user} deleted school '"+schoolName+"' with ID "+strconv.FormatInt(i.SchoolID, 10)+"."); err != nil {
		log.Println(err)
		return err
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
