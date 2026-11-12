package members

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"app/internal/permissions"
	"app/internal/schoollogs"
)

type DeleteStaffMemberInput struct {
	UserID  int64
	StaffID int64
}

func (s *Service) DeleteStaffMember(ctx context.Context, i *DeleteStaffMemberInput) error {
	userID := i.UserID
	targetStaffID := i.StaffID
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var schoolID int64
	var targetName string
	var targetEmail string

	err = tx.QueryRowContext(ctx, `
		SELECT ss.school_id, u.name, u.email
		FROM school_staff ss
		JOIN users u ON u.id = ss.user_id
		WHERE ss.id = $1
		FOR UPDATE
	`, targetStaffID).Scan(&schoolID, &targetName, &targetEmail)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		log.Println(err)
		return err
	}

	var isOwner bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM schools WHERE id = $1 AND owner_id = $2)
	`, schoolID, userID).Scan(&isOwner); err != nil {
		log.Println(err)
		return err
	}

	if !isOwner {
		var actorStaffID int64

		err = tx.QueryRowContext(ctx, `
			SELECT id
			FROM school_staff
			WHERE school_id = $1
			  AND user_id = $2
			FOR UPDATE
		`, schoolID, userID).Scan(&actorStaffID)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		}
		if err != nil {
			log.Println(err)
			return err
		}

		if actorStaffID == targetStaffID {
			return ErrForbidden
		}

		var targetHighestPos int

		err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(r.position), -1)
		FROM staff_role_members rm
		JOIN staff_roles r
			ON r.id = rm.role_id
		WHERE rm.staff_id = $1
		  AND r.school_id = $2
	`, targetStaffID, schoolID).Scan(&targetHighestPos)
		if err != nil {
			log.Println(err)
			return err
		}

		var allowed bool

		err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM staff_role_members rm
			JOIN staff_roles r
				ON r.id = rm.role_id
			JOIN staff_role_permissions rp
				ON rp.role_id = r.id
			WHERE rm.staff_id = $1
			  AND r.school_id = $2
			  AND r.position > $3
			  AND rp.permission = $4
		)
		`,
			actorStaffID,
			schoolID,
			targetHighestPos,
			permissions.PermissionStaffDelete,
		).Scan(&allowed)
		if err != nil {
			log.Println(err)
			return err
		}

		if !allowed {
			return ErrForbidden
		}
	}

	res, err := tx.ExecContext(ctx, `
		DELETE FROM school_staff
		WHERE id = $1
		  AND school_id = $2
	`, targetStaffID, schoolID)
	if err != nil {
		log.Println(err)
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return err
	}

	if affected == 0 {
		return ErrNotFound
	}

	if err := schoollogs.StoreSchoolLog(schoolID, userID, schoollogs.ActionStaffDelete, schoollogs.TypeDelete, "Staff member removed", "{user} removed "+targetName+" from the school staff.", tx, ctx, s.Sf, "{user} removed staff member "+targetName+" <"+targetEmail+">."); err != nil {
		log.Println(err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
