package students

import (
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"context"
	"log"
)

type UnassignGuardianInput struct {
	UserID     int64
	StudentID  int64
	GuardianID int64
}

func (s *Service) UnassignGuardian(ctx context.Context, i *UnassignGuardianInput) error {
	// Get the student & guardian school ID
	sSchoolID, err := staff_helpers.StudentToSchoolID(ctx, s.DB, i.StudentID)
	if err != nil {
		log.Println(err)
		return err
	}
	gSchoolID, err := staff_helpers.GuardianToSchoolID(ctx, s.DB, i.GuardianID)
	if err != nil {
		log.Println(err)
		return err
	}

	if sSchoolID != gSchoolID {
		return ErrForbidden
	}

	if !schools.Can(schools.PermissionStudentGuardianUnassign, i.UserID, sSchoolID, ctx, s.DB) {
		return ErrForbidden
	}

	if _, err := s.DB.ExecContext(ctx, `
		DELETE FROM guardian_students
		WHERE student_id = $1
		AND guardian_id = $2
	`,
		i.StudentID,
		i.GuardianID,
	); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
