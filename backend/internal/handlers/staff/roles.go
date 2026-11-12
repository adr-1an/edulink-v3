package staff

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"app/internal/application/staff/roles"
	"app/internal/helpers/staff"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) CreateRoleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get school ID
	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
	decodeErr := dec.Decode(&p)

	res, err := h.App.Staff.Roles.CreateRole(ctx, &roles.CreateRoleInput{
		InvalidPayload: decodeErr != nil,
		UserID:         userID,
		SchoolID:       schoolID,
		Role:           roles.CreateRole{Name: p.Name, Position: p.Position, Color: p.Color},
	})
	if err != nil {
		switch {
		case errors.Is(err, roles.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, roles.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, roles.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, roles.ErrConflict):
			w.WriteHeader(http.StatusConflict)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if res.NoRowsAffected {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) ListRolesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get school ID
	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	type Role struct {
		ID          string   `json:"id"`
		Position    int      `json:"position"`
		Name        string   `json:"name"`
		Color       string   `json:"color"`
		CreatedAt   string   `json:"createdAt"`
		Permissions []string `json:"permissions"`
	}
	res, err := h.App.Staff.Roles.ListRoles(ctx, &roles.ListRolesInput{UserID: userID, SchoolID: schoolID})
	if err != nil {
		switch {
		case errors.Is(err, roles.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	var roles []Role
	if res.Roles != nil {
		roles = make([]Role, 0, len(res.Roles))
	}
	for _, item := range res.Roles {
		roles = append(roles, Role{ID: item.ID, Position: item.Position, Name: item.Name, Color: item.Color, CreatedAt: item.CreatedAt, Permissions: item.Permissions})
	}
	access := res.Access

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access": access,
		"roles":  roles,
	})
}

func (h *Handler) UpdateRoleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get role ID
	roleID, err := strconv.ParseInt(chi.URLParam(r, "roleID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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
	decodeErr := dec.Decode(&p)

	res, err := h.App.Staff.Roles.UpdateRole(ctx, &roles.UpdateRoleInput{
		InvalidPayload: decodeErr != nil,
		UserID:         userID,
		RoleID:         roleID,
		Role:           roles.UpdateRole{Name: p.Name, Position: p.Position, Color: p.Color},
	})
	if err != nil {
		switch {
		case errors.Is(err, roles.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, roles.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, roles.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, roles.ErrConflict):
			w.WriteHeader(http.StatusConflict)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if res.NoRowsAffected {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteRoleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get role ID
	roleID, err := strconv.ParseInt(chi.URLParam(r, "roleID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Roles.DeleteRole(ctx, &roles.DeleteRoleInput{
		UserID: userID,
		RoleID: roleID,
	})
	if err != nil {
		switch {
		case errors.Is(err, roles.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if res.NoRowsAffected {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SetStaffRolePermissionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get role ID
	roleID, err := strconv.ParseInt(chi.URLParam(r, "roleID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
	decodeErr := dec.Decode(&p)

	err = h.App.Staff.Roles.SetStaffRolePermission(ctx, &roles.SetStaffRolePermissionInput{
		InvalidPayload: decodeErr != nil,
		UserID:         userID,
		RoleID:         roleID,
		Role:           roles.SetStaffRolePermission{Permission: p.Permission, Allow: p.Allow},
	})
	if err != nil {
		switch {
		case errors.Is(err, roles.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, roles.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, roles.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	_, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	res, err := h.App.Staff.Roles.ListPermissions(ctx, &roles.ListPermissionsInput{})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Return permissions
	_ = json.NewEncoder(w).Encode(map[string]any{
		"permissions": res.Permissions,
	})
}

func (h *Handler) AddStaffRoleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
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

	err = h.App.Staff.Roles.AddStaffRole(ctx, &roles.AddStaffRoleInput{
		UserID:  userID,
		StaffID: staffID,
		RoleID:  roleID,
	})
	if err != nil {
		switch {
		case errors.Is(err, roles.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveStaffRoleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
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

	err = h.App.Staff.Roles.RemoveStaffRole(ctx, &roles.RemoveStaffRoleInput{
		UserID:  userID,
		StaffID: staffID,
		RoleID:  roleID,
	})
	if err != nil {
		switch {
		case errors.Is(err, roles.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListStaffRolesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get the staff ID
	staffID, err := strconv.ParseInt(chi.URLParam(r, "staffID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	type Role struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Color    string `json:"color"`
		Position int    `json:"position"`
	}
	res, err := h.App.Staff.Roles.ListStaffRoles(ctx, &roles.ListStaffRolesInput{UserID: userID, StaffID: staffID})
	if err != nil {
		switch {
		case errors.Is(err, roles.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	var roles []Role
	if res.Roles != nil {
		roles = make([]Role, 0, len(res.Roles))
	}
	for _, item := range res.Roles {
		roles = append(roles, Role{ID: item.ID, Name: item.Name, Color: item.Color, Position: item.Position})
	}
	// Get user access data
	access := res.Access

	// Return data
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access": access,
		"roles":  roles,
	})
}
