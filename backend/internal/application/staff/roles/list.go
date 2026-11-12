package roles

import (
	"context"
	"log"

	"app/internal/helpers/staff/schools"

	"github.com/lib/pq"
)

type ListRole struct {
	ID          string
	Position    int
	Name        string
	Color       string
	CreatedAt   string
	Permissions []string
}
type ListRolesResult struct {
	Access   schools.AccessData
	Roles    []ListRole
	SchoolID int64
}
type ListRolesInput struct {
	UserID   int64
	SchoolID int64
}

func (s *Service) ListRoles(ctx context.Context, i *ListRolesInput) (*ListRolesResult, error) {
	userID := i.UserID
	schoolID := i.SchoolID
	// Check permissions
	if !schools.Can(schools.PermissionRoleList, userID, schoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	var roles []ListRole

	// Fetch roles
	rows, err := s.DB.QueryContext(ctx, `
		SELECT
			r.id,
			r.position,
			r.name,
			r.color,
			r.created_at,
			COALESCE(
				ARRAY_AGG(rp.permission ORDER BY rp.permission)
					FILTER (WHERE rp.permission IS NOT NULL),
				ARRAY[]::TEXT[]
			)
		FROM staff_roles r
		LEFT JOIN staff_role_permissions rp
			ON rp.role_id = r.id
		WHERE r.school_id = $1
		GROUP BY r.id, r.position, r.name, r.color, r.created_at
		ORDER BY r.position DESC
	`, schoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var r ListRole
		var permissions pq.StringArray

		if err := rows.Scan(&r.ID, &r.Position, &r.Name, &r.Color, &r.CreatedAt, &permissions); err != nil {
			log.Println(err)
			return nil, err
		}
		r.Permissions = permissions

		roles = append(roles, r)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	access, err := schools.GetAllUserPermissions(ctx, s.DB, userID, schoolID)
	if err != nil {
		return nil, err
	}

	return &ListRolesResult{Access: access, Roles: roles, SchoolID: schoolID}, nil
}
