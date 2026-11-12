package academicyears

import (
	"context"

	"app/internal/permissions"
)

type ListAcademicYear struct {
	ID        int64
	StartYear int
	EndYear   int
	IsActive  bool
}

type ListAcademicYearsInput struct {
	UserID   int64
	SchoolID int64
}

type ListAcademicYearsResult struct {
	Access        permissions.AccessData
	AcademicYears []ListAcademicYear
}

func (s *Service) ListAcademicYears(ctx context.Context, i *ListAcademicYearsInput) (*ListAcademicYearsResult, error) {
	if !permissions.Can(permissions.PermissionAcademicYearList, i.UserID, i.SchoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, start_year, end_year, is_active
		FROM academic_years
		WHERE school_id = $1
	`, i.SchoolID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var years []ListAcademicYear
	for rows.Next() {
		var y ListAcademicYear

		if err := rows.Scan(&y.ID, &y.StartYear, &y.EndYear, &y.IsActive); err != nil {
			return nil, err
		}

		years = append(years, y)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	access, err := permissions.GetAllUserPermissions(ctx, s.DB, i.UserID, i.SchoolID)
	if err != nil {
		return nil, err
	}

	return &ListAcademicYearsResult{Access: access,
		AcademicYears: years,
	}, nil
}
