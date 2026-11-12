package permissions

import (
	"context"
	"database/sql"
	"log"

	"github.com/lib/pq"
)

type Permission string

const (
	// The SchoolOwner permission is never granted through a role.
	// It exists so it can be passed into permission checks which only the owner is supposed to pass.
	SchoolOwner Permission = "owner"

	PermissionSchoolView    Permission = "school.view"
	PermissionSchoolUpdate  Permission = "school.update"
	PermissionSchoolPromote Permission = "school.promote"

	PermissionSchoolInviteList   Permission = "school.invite.list"
	PermissionSchoolInviteCancel Permission = "school.invite.cancel"

	PermissionStaffView       Permission = "staff.view"
	PermissionStaffCreate     Permission = "staff.create"
	PermissionStaffDelete     Permission = "staff.delete"
	PermissionStaffRoleAdd    Permission = "staff.role.add"
	PermissionStaffRoleRemove Permission = "staff.role.remove"
	PermissionStaffRoleList   Permission = "staff.role.list"

	PermissionAcademicYearCreate       Permission = "academicYear.create"
	PermissionAcademicYearList         Permission = "academicYear.list"
	PermissionAcademicYearToggleActive Permission = "academicYear.toggleActive"
	PermissionAcademicYearDelete       Permission = "academicYear.delete"

	PermissionGradeList   Permission = "grade.list"
	PermissionGradeCreate Permission = "grade.create"
	PermissionGradeUpdate Permission = "grade.update"
	PermissionGradeDelete Permission = "grade.delete"

	PermissionRoleCreate Permission = "role.create"
	PermissionRoleList   Permission = "role.list"
	PermissionRoleUpdate Permission = "role.update"
	PermissionRoleDelete Permission = "role.delete"

	PermissionRolePermissionUpdate Permission = "role.permission.update"

	PermissionCourseCreate Permission = "course.create"
	PermissionCourseList   Permission = "course.list"
	PermissionCourseUpdate Permission = "course.update"
	PermissionCourseDelete Permission = "course.delete"

	PermissionCoursePostCreate Permission = "course.post.create"
	PermissionCoursePostList   Permission = "course.post.list"
	PermissionCoursePostView   Permission = "course.post.view"
	PermissionCoursePostUpdate Permission = "course.post.update"
	PermissionCoursePostDelete Permission = "course.post.delete"

	PermissionPostAttachmentCreate Permission = "post.attachment.create"
	PermissionPostAttachmentDelete Permission = "post.attachment.delete"

	PermissionLogsList Permission = "log.list"

	PermissionStudentCreate           Permission = "student.create"
	PermissionStudentList             Permission = "student.list"
	PermissionStudentUpdate           Permission = "student.update"
	PermissionStudentDelete           Permission = "student.delete"
	PermissionStudentView             Permission = "student.view"
	PermissionStudentGuardianAssign   Permission = "student.guardian.assign"
	PermissionStudentGuardianUnassign Permission = "student.guardian.unassign"
	PermissionStudentGuardianList     Permission = "student.guardian.list"

	PermissionGuardianCreate Permission = "guardian.create"
	PermissionGuardianList   Permission = "guardian.list"
	PermissionGuardianUpdate Permission = "guardian.update"
	PermissionGuardianDelete Permission = "guardian.delete"
	PermissionGuardianView   Permission = "guardian.view"

	PermissionGuardianStudentList Permission = "guardian.student.list"

	PermissionPortalUserActivate   Permission = "portal.user.activate"
	PermissionPortalUserDeactivate Permission = "portal.user.deactivate"

	PermissionCourseStudentAssign Permission = "course.student.assign"
	PermissionCourseStudentRemove Permission = "course.student.remove"
	PermissionCourseStudentList   Permission = "course.student.list"

	PermissionCourseAssignmentCreate Permission = "course.assignment.create"
	PermissionCourseAssignmentList   Permission = "course.assignment.list"
	PermissionCourseAssignmentUpdate Permission = "course.assignment.update"
	PermissionCourseAssignmentDelete Permission = "course.assignment.delete"

	PermissionSubmissionList        Permission = "submission.list"
	PermissionSubmissionView        Permission = "submission.view"
	PermissionSubmissionReturn      Permission = "submission.return"
	PermissionSubmissionDelete      Permission = "submission.delete"
	PermissionSubmissionGrade       Permission = "submission.grade"
	PermissionSubmissionRemoveGrade Permission = "submission.removeGrade"
)

func (p Permission) String() string {
	return string(p)
}

func PortalUserActiveStatePermission(wasActive, isActive bool) (Permission, bool) {
	if wasActive == isActive {
		return "", false
	}
	if isActive {
		return PermissionPortalUserActivate, true
	}
	return PermissionPortalUserDeactivate, true
}

var ValidPermissions = map[Permission]struct{}{
	PermissionSchoolView:         {},
	PermissionSchoolUpdate:       {},
	PermissionSchoolPromote:      {},
	PermissionSchoolInviteList:   {},
	PermissionSchoolInviteCancel: {},

	PermissionStaffView:       {},
	PermissionStaffCreate:     {},
	PermissionStaffDelete:     {},
	PermissionStaffRoleAdd:    {},
	PermissionStaffRoleRemove: {},
	PermissionStaffRoleList:   {},

	PermissionAcademicYearCreate:       {},
	PermissionAcademicYearList:         {},
	PermissionAcademicYearToggleActive: {},
	PermissionAcademicYearDelete:       {},

	PermissionGradeList:   {},
	PermissionGradeCreate: {},
	PermissionGradeUpdate: {},
	PermissionGradeDelete: {},

	PermissionRoleCreate: {},
	PermissionRoleList:   {},
	PermissionRoleUpdate: {},
	PermissionRoleDelete: {},

	PermissionRolePermissionUpdate: {},

	PermissionCourseCreate:   {},
	PermissionCourseUpdate:   {},
	PermissionCourseList:     {},
	PermissionCoursePostView: {},
	PermissionCourseDelete:   {},

	PermissionCoursePostCreate: {},
	PermissionCoursePostList:   {},
	PermissionCoursePostUpdate: {},
	PermissionCoursePostDelete: {},

	PermissionPostAttachmentCreate: {},
	PermissionPostAttachmentDelete: {},

	PermissionLogsList: {},

	PermissionStudentCreate:           {},
	PermissionStudentList:             {},
	PermissionStudentUpdate:           {},
	PermissionStudentDelete:           {},
	PermissionStudentView:             {},
	PermissionStudentGuardianAssign:   {},
	PermissionStudentGuardianUnassign: {},
	PermissionStudentGuardianList:     {},

	PermissionGuardianCreate: {},
	PermissionGuardianList:   {},
	PermissionGuardianUpdate: {},
	PermissionGuardianDelete: {},
	PermissionGuardianView:   {},

	PermissionGuardianStudentList: {},

	PermissionPortalUserActivate:   {},
	PermissionPortalUserDeactivate: {},

	PermissionCourseStudentAssign: {},
	PermissionCourseStudentRemove: {},
	PermissionCourseStudentList:   {},

	PermissionCourseAssignmentCreate: {},
	PermissionCourseAssignmentList:   {},
	PermissionCourseAssignmentUpdate: {},
	PermissionCourseAssignmentDelete: {},

	PermissionSubmissionList: {},
	PermissionSubmissionView: {},

	PermissionSubmissionReturn:      {},
	PermissionSubmissionDelete:      {},
	PermissionSubmissionGrade:       {},
	PermissionSubmissionRemoveGrade: {},
}

func ValidatePermissionString(s string) bool {
	_, ok := ValidPermissions[Permission(s)]
	return ok
}

type QueryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func Can(action Permission, staffID, schoolID int64, ctx context.Context, db *sql.DB) bool {
	// Check if the user owns the school or has the required permission
	var allowed bool
	err := db.QueryRowContext(ctx, `
	SELECT EXISTS (
	    SELECT 1 FROM schools WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL
	)
	OR EXISTS (
	    SELECT 1 FROM school_staff ss
	 	JOIN staff_role_members srm
	 		ON srm.staff_id = ss.id
	 	JOIN staff_roles sr
	 		ON sr.id = srm.role_id
	 	JOIN staff_role_permissions srp
	 		ON srp.role_id = sr.id
	 	JOIN schools s
		 	ON s.id = sr.school_id
	 	WHERE ss.user_id = $2
	 	AND ss.school_id = $1
	 	AND sr.school_id = $1
	 	AND srp.permission = $3
	 	AND s.deleted_at IS NULL
	)
`, schoolID, staffID, action).Scan(&allowed)
	if err != nil {
		log.Println(err)
		return false
	}

	return allowed
}

func OwnsSchoolOrHasRoleAbovePosition(ctx context.Context, q QueryRower, userID, schoolID int64, position int) (bool, error) {
	var allowed bool
	err := q.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM schools
			WHERE id = $1
			  AND owner_id = $2
			  AND deleted_at IS NULL
		)
		OR COALESCE((
			SELECT MAX(sr.position)
			FROM school_staff ss
			JOIN staff_role_members srm
				ON srm.staff_id = ss.id
			JOIN staff_roles sr
				ON sr.id = srm.role_id
			WHERE ss.school_id = $1
			  AND ss.user_id = $2
			  AND sr.school_id = $1
		), -1) > $3
	`, schoolID, userID, position).Scan(&allowed)
	if err != nil {
		return false, err
	}

	return allowed, nil
}

func GetAllUserPermissions(ctx context.Context, db *sql.DB, userID, schoolID int64) (AccessData, error) {
	rows, err := db.QueryContext(ctx, `
		WITH owner_check AS (
			SELECT EXISTS (
				SELECT 1
				FROM schools
				WHERE id = $1
				  AND owner_id = $2
				  AND deleted_at IS NULL
			) AS is_owner
		)
		SELECT
			oc.is_owner,
			sr.position,
			COALESCE(
				ARRAY_AGG(srp.permission ORDER BY srp.permission)
					FILTER (WHERE srp.permission IS NOT NULL),
				ARRAY[]::TEXT[]
			) AS permissions
		FROM owner_check oc
		LEFT JOIN school_staff ss
			ON ss.user_id = $2
		   AND ss.school_id = $1
		LEFT JOIN staff_role_members srm
			ON srm.staff_id = ss.id
		LEFT JOIN staff_roles sr
			ON sr.id = srm.role_id
		   AND sr.school_id = $1
		LEFT JOIN staff_role_permissions srp
			ON srp.role_id = sr.id
		GROUP BY oc.is_owner, sr.id, sr.position
		ORDER BY sr.position DESC
	`, schoolID, userID)
	if err != nil {
		return AccessData{}, err
	}
	defer func() { _ = rows.Close() }()

	access := AccessData{
		Roles: make([]struct {
			Position    int      `json:"position"`
			Permissions []string `json:"permissions"`
		}, 0),
	}

	for rows.Next() {
		var isOwner bool
		var position *int64
		var permissions pq.StringArray

		if err := rows.Scan(&isOwner, &position, &permissions); err != nil {
			return AccessData{}, err
		}

		access.Owner = isOwner
		if position == nil {
			continue
		}

		access.Roles = append(access.Roles, struct {
			Position    int      `json:"position"`
			Permissions []string `json:"permissions"`
		}{
			Position:    int(*position),
			Permissions: permissions,
		})
	}

	if err := rows.Err(); err != nil {
		return AccessData{}, err
	}

	return access, nil
}

type AccessData struct {
	Owner bool `json:"owner"`
	Roles []struct {
		Position    int      `json:"position"`
		Permissions []string `json:"permissions"`
	} `json:"roles"`
}
