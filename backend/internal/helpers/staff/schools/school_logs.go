package schools

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/sony/sonyflake/v2"
)

type Action string
type Type string

const (
	ActionSchoolCreate Action = "school.create"
	ActionSchoolEdit   Action = "school.edit"
	ActionSchoolDelete Action = "school.delete"

	ActionAcademicYearCreate Action = "academicYear.create"
	ActionAcademicYearDelete Action = "academicYear.delete"

	ActionGradeCreate Action = "grade.create"
	ActionGradeEdit   Action = "grade.edit"
	ActionGradeDelete Action = "grade.delete"

	ActionCourseCreate Action = "course.create"
	ActionCourseEdit   Action = "course.edit"
	ActionCourseDelete Action = "course.delete"

	ActionPostCreate Action = "post.create"
	ActionPostEdit   Action = "post.edit"
	ActionPostDelete Action = "post.delete"

	ActionRoleCreate            Action = "role.create"
	ActionRoleEdit              Action = "role.edit"
	ActionRoleDelete            Action = "role.delete"
	ActionRolePermissionEdit    Action = "rolePermission.edit"
	ActionStaffCreate           Action = "staff.create"
	ActionStaffDelete           Action = "staff.delete"
	ActionStaffRoleCreate       Action = "staffRole.create"
	ActionStaffRoleDelete       Action = "staffRole.delete"
	ActionStaffInvitationCreate Action = "staffInvitation.create"
	ActionStaffInvitationEdit   Action = "staffInvitation.edit"
	ActionStaffInvitationDelete Action = "staffInvitation.delete"
	ActionSchoolPromote         Action = "school.promote"
	ActionSchoolLeave           Action = "school.leave"

	ActionStaffInvitationCancel Action = "staffInvitation.cancel"

	ActionStudentCreate Action = "student.create"
	ActionStudentEdit   Action = "student.edit"
	ActionStudentDelete Action = "student.delete"

	ActionCourseStudentAssign Action = "course.student.assign"
	ActionCourseStudentRemove Action = "course.student.remove"

	ActionAssignmentCreate Action = "assignment.create"
	ActionAssignmentEdit   Action = "assignment.edit"
	ActionAssignmentDelete Action = "assignment.delete"
)

const (
	TypeCreate Type = "create"
	TypeEdit   Type = "edit"
	TypeDelete Type = "delete"
	TypeOther  Type = "other"
)

func StoreSchoolLog(
	schoolID, userID int64,
	action Action,
	actionType Type,
	title, message string,
	tx *sql.Tx,
	ctx context.Context,
	sf *sonyflake.Sonyflake,
	details ...string,
) error {
	id, err := sf.NextID()
	if err != nil {
		return err
	}

	var userName string
	if err := tx.QueryRowContext(ctx, `
		SELECT name FROM users WHERE id = $1
	`, userID).Scan(&userName); err != nil {
		return err
	}

	message = strings.ReplaceAll(message, "{user}", userName)
	logDetails := message
	if len(details) > 0 {
		logDetails = strings.ReplaceAll(details[0], "{user}", userName)
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO school_logs (id, school_id, by_user, type, action, title, message, details)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, schoolID, userID, actionType, action, title, message, logDetails)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("0 rows were affected")
	}

	return nil
}
