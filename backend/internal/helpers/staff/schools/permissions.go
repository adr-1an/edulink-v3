package schools

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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

	PermissionStudentCreate Permission = "student.create"
	PermissionStudentList   Permission = "student.list"
	PermissionStudentUpdate Permission = "student.update"
	PermissionStudentDelete Permission = "student.delete"
	PermissionStudentView   Permission = "student.view"

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

	PermissionStudentCreate: {},
	PermissionStudentList:   {},
	PermissionStudentUpdate: {},
	PermissionStudentDelete: {},
	PermissionStudentView:   {},

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

// YearToSchoolID returns the ID of the school that the Academic Year belongs to.
func YearToSchoolID(db *sql.DB, w http.ResponseWriter, r *http.Request, ctx context.Context) (int64, int64, error) {
	yearIDStr := chi.URLParam(r, "yearID")
	yearID, err := strconv.ParseInt(yearIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return 0, 0, errors.New("yearID is not a number")
	}

	var schoolID int64
	if err := db.QueryRowContext(ctx, `
		SELECT
		    school_id
		FROM academic_years
		WHERE id = $1
	`, yearID).Scan(&schoolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return 0, 0, err
	}

	return schoolID, yearID, nil
}

// GradeToSchoolID returns the ID of the school that the Grade belongs to.
func GradeToSchoolID(db *sql.DB, w http.ResponseWriter, r *http.Request, ctx context.Context) (int64, int64, error) {
	gradeIDStr := chi.URLParam(r, "gradeID")
	gradeID, err := strconv.ParseInt(gradeIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return 0, 0, errors.New("gradeID is not a number")
	}

	var schoolID int64
	if err := db.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM academic_years y
		JOIN grades g ON g.academic_year_id = y.id
		WHERE g.id = $1
	`, gradeID).Scan(&schoolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return 0, 0, err
	}

	return schoolID, gradeID, nil
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
		var position sql.NullInt64
		var permissions pq.StringArray

		if err := rows.Scan(&isOwner, &position, &permissions); err != nil {
			return AccessData{}, err
		}

		access.Owner = isOwner
		if !position.Valid {
			continue
		}

		access.Roles = append(access.Roles, struct {
			Position    int      `json:"position"`
			Permissions []string `json:"permissions"`
		}{
			Position:    int(position.Int64),
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
