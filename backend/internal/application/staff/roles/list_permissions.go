package roles

import (
	"context"
	"sort"

	"app/internal/helpers/staff/schools"
)

type ListPermissionsInput struct{}
type ListPermissionsResult struct{ Permissions []string }

func (s *Service) ListPermissions(ctx context.Context, i *ListPermissionsInput) (*ListPermissionsResult, error) {
	// Get all valid permissions
	var perms []string

	for perm := range schools.ValidPermissions {
		p := perm.String()
		perms = append(perms, p)
	}
	sort.Strings(perms)

	return &ListPermissionsResult{Permissions: perms}, nil
}
