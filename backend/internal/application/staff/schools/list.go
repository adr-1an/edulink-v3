package schools

import (
	"context"
	"log"
)

type ListSchoolsInput struct {
	UserID int64
}

type ListSchoolsResult struct {
	Schools []School
}

type School struct {
	ID         int64
	OwnerID    int64
	Name       string
	RegionCode string
}

func (s *Service) ListSchools(ctx context.Context, i *ListSchoolsInput) (*ListSchoolsResult, error) {
	var res ListSchoolsResult

	rows, err := s.DB.QueryContext(ctx, `
		SELECT DISTINCT s.id, s.owner_id, s.name, s.region_code
		FROM schools s
		WHERE s.deleted_at IS NULL
  		AND (
    		s.owner_id = $1
    		OR EXISTS (
				SELECT 1
      			FROM school_staff ss
      			WHERE ss.user_id = $1
        		AND ss.school_id = s.id
        		AND s.deleted_at IS NULL
    		)
  		);
	`, i.UserID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var school School

		if err := rows.Scan(&school.ID, &school.OwnerID, &school.Name, &school.RegionCode); err != nil {
			log.Println(err)
			return nil, err
		}

		res.Schools = append(res.Schools, school)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &res, nil
}
