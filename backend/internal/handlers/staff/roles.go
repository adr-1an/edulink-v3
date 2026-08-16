package staff

import (
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"github.com/sony/sonyflake/v2"
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

func CreateRoleHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get school ID
	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionRoleCreate, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Payload
	type Payload struct {
		Name     string  `json:"name"`
		Position int     `json:"position"`
		Color    *string `json:"color"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse
	p.Name = strings.TrimSpace(p.Name)
	var color string
	if p.Color != nil {
		color = *p.Color
	}

	// Validate
	if p.Name == "" || len(p.Name) > 32 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if len(color) != 6 { // invalid hex len
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	order, err := lockSchoolRoleOrder(ctx, tx, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if p.Position < 0 || p.Position > len(order) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Check if the actor can create a role at this position, or if the actor is the school owner.
	allowed, err := schools.OwnsSchoolOrHasRoleAbovePosition(ctx, tx, userID, schoolID, p.Position)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !allowed {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Check for a name conflict. Occupied positions are shifted automatically.
	var conflict bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM staff_roles
			WHERE name = $1
			AND school_id = $2
		)
	`, p.Name, schoolID).Scan(&conflict); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if conflict {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// Generate ID
	id, err := sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var temporaryPosition int
	if len(order) > 0 {
		temporaryPosition = order[len(order)-1].Position + 1
	}

	// Create the role in a free temporary slot, then rank every role contiguously.
	res, err := tx.ExecContext(ctx, `
		INSERT INTO staff_roles (id, position, school_id, created_by, name, color)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, temporaryPosition, schoolID, userID, p.Name, color)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		log.Println(err)
		return
	}

	orderedRoleIDs := insertRoleID(roleIDs(order), id, p.Position)
	if err := persistRoleOrder(ctx, tx, schoolID, orderedRoleIDs); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Log action
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionRoleCreate, schools.TypeCreate, "Role created", "{user} created the role '"+p.Name+"'.", tx, ctx, sf, "{user} created role '"+p.Name+"' at position "+strconv.Itoa(p.Position)+" with color '#"+color+"'."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func ListRolesHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get school ID
	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionRoleList, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Role struct
	type Role struct {
		ID          string   `json:"id"`
		Position    int      `json:"position"`
		Name        string   `json:"name"`
		Color       string   `json:"color"`
		CreatedAt   string   `json:"createdAt"`
		Permissions []string `json:"permissions"`
	}
	var roles []Role

	// Fetch roles
	rows, err := db.QueryContext(ctx, `
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var r Role
		var permissions pq.StringArray

		if err := rows.Scan(&r.ID, &r.Position, &r.Name, &r.Color, &r.CreatedAt, &permissions); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Permissions = permissions

		roles = append(roles, r)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"roles": roles,
	})
}

func UpdateRoleHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get role ID
	roleID, err := strconv.ParseInt(chi.URLParam(r, "roleID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the role's school ID
	var schoolID int64
	var roleName string
	if err := db.QueryRowContext(ctx, `
		SELECT school_id, name FROM staff_roles WHERE id = $1
	`, roleID).Scan(&schoolID, &roleName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionRoleUpdate, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Payload
	type Payload struct {
		Name     string `json:"name"`
		Color    string `json:"color"`
		Position int    `json:"position"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse
	p.Name = strings.TrimSpace(p.Name)

	// Validate
	if p.Name == "" || len(p.Name) > 32 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if len(p.Color) != 6 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	order, err := lockSchoolRoleOrder(ctx, tx, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if p.Position < 0 || p.Position >= len(order) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	currentIndex := -1
	currentPosition := -1
	for i, role := range order {
		if role.ID == roleID {
			currentIndex = i
			currentPosition = role.Position
			break
		}
	}
	if currentIndex == -1 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var oldName string
	var oldColor string
	if err := tx.QueryRowContext(ctx, `
		SELECT name, color
		FROM staff_roles
		WHERE id = $1
		AND school_id = $2
	`, roleID, schoolID).Scan(&oldName, &oldColor); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check if the actor can update the target role and move it to this position, or if the actor is the school owner
	canUpdateCurrent, err := schools.OwnsSchoolOrHasRoleAbovePosition(ctx, tx, userID, schoolID, currentPosition)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	canSetPosition, err := schools.OwnsSchoolOrHasRoleAbovePosition(ctx, tx, userID, schoolID, p.Position)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !canUpdateCurrent || !canSetPosition {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Check for a name conflict. Moving to an occupied rank shifts the roles in between.
	var conflict bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM staff_roles
			WHERE school_id = $1
			  AND id <> $2
			  AND name = $3
		)
	`, schoolID, roleID, p.Name).Scan(&conflict); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if conflict {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// Update the role fields, then atomically rebuild the contiguous role order.
	res, err := tx.ExecContext(ctx, `
		UPDATE staff_roles
		SET name = $1,
		    color = $2,
		    updated_at = NOW()
		WHERE id = $3
		  AND school_id = $4
	`, p.Name, p.Color, roleID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		log.Println(err)
		return
	}

	orderedRoleIDs := moveRoleID(roleIDs(order), currentIndex, p.Position)
	if err := persistRoleOrder(ctx, tx, schoolID, orderedRoleIDs); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Log action
	details := "{user} updated the role name from '" + oldName + "' to '" + p.Name + "', color from '#" + oldColor + "' to '#" + p.Color + "', and position from " + strconv.Itoa(currentPosition) + " to " + strconv.Itoa(p.Position) + "."
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionRoleEdit, schools.TypeEdit, "Role updated", "{user} updated the role '"+p.Name+"'.", tx, ctx, sf, details); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteRoleHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get role ID
	roleID, err := strconv.ParseInt(chi.URLParam(r, "roleID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the role's school ID
	var schoolID int64
	var roleName string
	if err := db.QueryRowContext(ctx, `
		SELECT school_id, name FROM staff_roles WHERE id = $1
	`, roleID).Scan(&schoolID, &roleName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionRoleDelete, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	order, err := lockSchoolRoleOrder(ctx, tx, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	deleteIndex := -1
	currentPosition := -1
	for i, role := range order {
		if role.ID == roleID {
			deleteIndex = i
			currentPosition = role.Position
			break
		}
	}
	if deleteIndex == -1 {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Check if the actor can delete this role, or if the actor is the school owner
	allowed, err := schools.OwnsSchoolOrHasRoleAbovePosition(ctx, tx, userID, schoolID, currentPosition)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !allowed {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Delete role
	res, err := tx.ExecContext(ctx, `
		DELETE FROM staff_roles
		WHERE id = $1
		  AND school_id = $2
	`, roleID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		log.Println(err)
		return
	}

	orderedRoleIDs := roleIDs(order)
	orderedRoleIDs = append(orderedRoleIDs[:deleteIndex], orderedRoleIDs[deleteIndex+1:]...)
	if err := persistRoleOrder(ctx, tx, schoolID, orderedRoleIDs); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Log action
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionRoleDelete, schools.TypeDelete, "Role deleted", "{user} deleted the role '"+roleName+"'.", tx, ctx, sf, "{user} deleted role '"+roleName+"' with ID "+strconv.FormatInt(roleID, 10)+"."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func SetStaffRolePermissionHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get role ID
	roleID, err := strconv.ParseInt(chi.URLParam(r, "roleID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the role's school ID and position
	var position int
	var schoolID int64
	var roleName string
	if err := db.QueryRowContext(ctx, `
		SELECT position, school_id, name
		FROM staff_roles
		WHERE id = $1
	`, roleID).Scan(&position, &schoolID, &roleName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionRolePermissionUpdate, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Payload
	type Payload struct {
		Permission string `json:"permission"`
		Allow      bool   `json:"allow"`
	}
	var p Payload

	/*
		If p.Allow is true, it'll insert a new permission row.
		If it's false, it'll attempt to delete the permission row.
		A permission is OFF by default unless the row exists.
	*/

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse
	p.Permission = strings.TrimSpace(p.Permission)

	// Validate
	if !schools.ValidatePermissionString(p.Permission) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Check if the user owns the school or can edit the role
	allowed, err := schools.OwnsSchoolOrHasRoleAbovePosition(ctx, tx, userID, schoolID, position)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !allowed {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if p.Allow {
		// insert the permission
		if _, err := tx.ExecContext(ctx, `
		INSERT INTO staff_role_permissions (role_id, permission)
		VALUES ($1, $2)
		ON CONFLICT (role_id, permission) DO NOTHING
	`, roleID, p.Permission); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		// delete the permission
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM staff_role_permissions rp
			USING staff_roles r
			WHERE rp.role_id = r.id
			AND rp.permission = $1
			AND r.school_id = $2
			AND rp.role_id = $3
		`, p.Permission, schoolID, roleID); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	title := "Role permission revoked"
	message := "{user} revoked the permission '" + p.Permission + "' from the role '" + roleName + "'."
	if p.Allow {
		title = "Role permission granted"
		message = "{user} granted the permission '" + p.Permission + "' to the role '" + roleName + "'."
	}

	// Log action
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionRolePermissionEdit, schools.TypeEdit, title, message, tx, ctx, sf); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ListPermissionsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	// Get user ID
	_, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get all valid permissions
	var perms []string

	for perm := range schools.ValidPermissions {
		p := perm.String()
		perms = append(perms, p)
	}
	sort.Strings(perms)

	// Return permissions
	_ = json.NewEncoder(w).Encode(map[string]any{
		"permissions": perms,
	})
}

func AddStaffRoleHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get staff ID & role ID
	staffID, err := strconv.ParseInt(chi.URLParam(r, "staffID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	roleID, err := strconv.ParseInt(chi.URLParam(r, "roleID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Get staff and role data
	target, err := getStaffRoleTarget(ctx, tx, staffID, roleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if target.StaffSchoolID != target.RoleSchoolID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionStaffRoleAdd, userID, target.StaffSchoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Check if the user can assign this role
	allowed, err := schools.OwnsSchoolOrHasRoleAbovePosition(ctx, tx, userID, target.StaffSchoolID, target.RolePosition)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !allowed {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Add role to staff member
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO staff_role_members (staff_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, staffID, roleID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Log action
	if err := schools.StoreSchoolLog(target.StaffSchoolID, userID, schools.ActionStaffRoleCreate, schools.TypeCreate, "Staff role added", "{user} added the role '"+target.RoleName+"' to "+target.StaffName+".", tx, ctx, sf, "{user} added role '"+target.RoleName+"' to staff member "+target.StaffName+"."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func RemoveStaffRoleHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get staff ID & role ID
	staffID, err := strconv.ParseInt(chi.URLParam(r, "staffID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	roleID, err := strconv.ParseInt(chi.URLParam(r, "roleID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Get staff and role data
	target, err := getStaffRoleTarget(ctx, tx, staffID, roleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if target.StaffSchoolID != target.RoleSchoolID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionStaffRoleRemove, userID, target.StaffSchoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Check if the user can remove this role
	allowed, err := schools.OwnsSchoolOrHasRoleAbovePosition(ctx, tx, userID, target.StaffSchoolID, target.RolePosition)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !allowed {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Remove role from staff member
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM staff_role_members
		WHERE staff_id = $1
		AND role_id = $2
	`, staffID, roleID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Log action
	if err := schools.StoreSchoolLog(target.StaffSchoolID, userID, schools.ActionStaffRoleDelete, schools.TypeDelete, "Staff role removed", "{user} removed the role '"+target.RoleName+"' from "+target.StaffName+".", tx, ctx, sf, "{user} removed role '"+target.RoleName+"' from staff member "+target.StaffName+"."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ListStaffRolesHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get the staff ID
	staffID, err := strconv.ParseInt(chi.URLParam(r, "staffID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the staff member's school ID
	var schoolID int64
	if err := db.QueryRowContext(ctx, `
		SELECT school_id FROM school_staff WHERE id = $1
	`, staffID).Scan(&schoolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionStaffRoleList, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Role struct
	type Role struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Color    string `json:"color"`
		Position int    `json:"position"`
	}
	var roles []Role

	// Fetch roles
	rows, err := db.QueryContext(ctx, `
		SELECT sr.id, sr.name, sr.color, sr.position
		FROM staff_roles sr
		JOIN staff_role_members srm
		ON srm.role_id = sr.id
		WHERE srm.staff_id = $1
		AND sr.school_id = $2
		ORDER BY sr.position DESC
	`, staffID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var r Role

		if err := rows.Scan(
			&r.ID,
			&r.Name,
			&r.Color,
			&r.Position,
		); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		roles = append(roles, r)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get user access data
	access, err := schools.GetAllUserPermissions(ctx, db, userID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return data
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access": access,
		"roles":  roles,
	})
}
