package students

import (
	"context"
	"log"

	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
)

type AssignGuardianInput struct {
	UserID     int64
	StudentID  int64
	GuardianID int64
}

func (s *Service) AssignGuardian(ctx context.Context, i *AssignGuardianInput) error {
	// Get the student & guardian school ID
	sSchoolID, err := staffhelpers.StudentToSchoolID(ctx, s.DB, i.StudentID)
	if err != nil {
		log.Println(err)
		return err
	}
	gSchoolID, err := staffhelpers.GuardianToSchoolID(ctx, s.DB, i.GuardianID)
	if err != nil {
		log.Println(err)
		return err
	}

	if sSchoolID != gSchoolID {
		return ErrForbidden
	}

	if !schools.Can(schools.PermissionStudentGuardianAssign, i.UserID, sSchoolID, ctx, s.DB) {
		return ErrForbidden
	}

	if _, err := s.DB.ExecContext(ctx, `
		INSERT INTO guardian_students (student_id, guardian_id, created_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (guardian_id, student_id) DO NOTHING
	`,
		i.StudentID,
		i.GuardianID,
		i.UserID,
	); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
