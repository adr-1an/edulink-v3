package roles

import (
	"context"
	"database/sql"
	"fmt"
)

type roleRank struct {
	ID       int64
	Position int
}

type staffRoleTarget struct {
	StaffSchoolID int64
	RoleSchoolID  int64
	RolePosition  int
	StaffName     string
	RoleName      string
}

func getStaffRoleTarget(ctx context.Context, tx *sql.Tx, staffID, roleID int64) (staffRoleTarget, error) {
	var target staffRoleTarget

	if err := tx.QueryRowContext(ctx, `
		SELECT ss.school_id, sr.school_id, sr.position, u.name, sr.name
		FROM school_staff ss
		JOIN users u ON u.id = ss.user_id
		CROSS JOIN staff_roles sr
		WHERE ss.id = $1
		AND sr.id = $2
	`, staffID, roleID).Scan(
		&target.StaffSchoolID,
		&target.RoleSchoolID,
		&target.RolePosition,
		&target.StaffName,
		&target.RoleName,
	); err != nil {
		return staffRoleTarget{}, err
	}

	return target, nil
}

func lockSchoolRoleOrder(ctx context.Context, tx *sql.Tx, schoolID int64) ([]roleRank, error) {
	var lockedSchoolID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM schools
		WHERE id = $1
		FOR UPDATE
	`, schoolID).Scan(&lockedSchoolID); err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id, position
		FROM staff_roles
		WHERE school_id = $1
		ORDER BY position, id
		FOR UPDATE
	`, schoolID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var order []roleRank
	for rows.Next() {
		var role roleRank
		if err := rows.Scan(&role.ID, &role.Position); err != nil {
			return nil, err
		}
		order = append(order, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return order, nil
}

func roleIDs(order []roleRank) []int64 {
	ids := make([]int64, len(order))
	for i, role := range order {
		ids[i] = role.ID
	}
	return ids
}

func insertRoleID(ids []int64, roleID int64, position int) []int64 {
	result := make([]int64, 0, len(ids)+1)
	result = append(result, ids[:position]...)
	result = append(result, roleID)
	result = append(result, ids[position:]...)
	return result
}

func moveRoleID(ids []int64, from, to int) []int64 {
	roleID := ids[from]
	withoutRole := make([]int64, 0, len(ids)-1)
	withoutRole = append(withoutRole, ids[:from]...)
	withoutRole = append(withoutRole, ids[from+1:]...)
	return insertRoleID(withoutRole, roleID, to)
}

func persistRoleOrder(ctx context.Context, tx *sql.Tx, schoolID int64, orderedRoleIDs []int64) error {
	if len(orderedRoleIDs) == 0 {
		return nil
	}

	var maxPosition int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(position), -1)
		FROM staff_roles
		WHERE school_id = $1
	`, schoolID).Scan(&maxPosition); err != nil {
		return err
	}

	offset := maxPosition + int64(len(orderedRoleIDs)) + 1
	const maxPostgresInt = int64(^uint32(0) >> 1)
	if maxPosition+offset > maxPostgresInt {
		return fmt.Errorf("role positions exceed PostgreSQL integer range")
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE staff_roles
		SET position = position + $1
		WHERE school_id = $2
	`, offset, schoolID); err != nil {
		return err
	}

	for position, roleID := range orderedRoleIDs {
		res, err := tx.ExecContext(ctx, `
			UPDATE staff_roles
			SET position = $1,
			    updated_at = CASE
			        WHEN position - $4 <> $1 THEN NOW()
			        ELSE updated_at
			    END
			WHERE id = $2
			  AND school_id = $3
		`, position, roleID, schoolID, offset)
		if err != nil {
			return err
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected != 1 {
			return fmt.Errorf("expected to rank one role, ranked %d", rowsAffected)
		}
	}

	return nil
}
