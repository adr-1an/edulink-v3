package staff

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"app/internal/application/staff/members"
	"app/internal/auth"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListStaffMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get school ID
	schoolIDStr := chi.URLParam(r, "schoolID")
	schoolID, err := strconv.ParseInt(schoolIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	type UserSummary struct {
		ID     string  `json:"id"`
		Name   string  `json:"name"`
		Email  string  `json:"email"`
		PfpURL *string `json:"profilePictureURL"`
	}
	type StaffRole struct {
		ID       string `json:"id"`
		Position int    `json:"position"`
		Name     string `json:"name"`
		Color    string `json:"color"`
	}
	type Staff struct {
		ID        string      `json:"id"`
		UserID    string      `json:"userId"`
		User      UserSummary `json:"user"`
		AddedBy   UserSummary `json:"addedBy"`
		CreatedAt time.Time   `json:"createdAt"`
		Roles     []StaffRole `json:"roles"`
	}
	res, err := h.App.Staff.Members.ListStaffMembers(ctx, &members.ListStaffMembersInput{UserID: userID, SchoolID: schoolID})
	if err != nil {
		switch {
		case errors.Is(err, members.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	var sMembers []Staff
	if res.Staff != nil {
		sMembers = make([]Staff, 0, len(res.Staff))
	}
	for _, item := range res.Staff {
		sMembers = append(sMembers, Staff{ID: item.ID, UserID: item.UserID, User: UserSummary{ID: item.User.ID, Name: item.User.Name, Email: item.User.Email, PfpURL: item.User.PfpURL}, AddedBy: UserSummary{ID: item.AddedBy.ID, Name: item.AddedBy.Name, Email: item.AddedBy.Email, PfpURL: item.AddedBy.PfpURL}, CreatedAt: item.CreatedAt, Roles: func() []StaffRole {
			if item.Roles == nil {
				return nil
			}
			items := make([]StaffRole, 0, len(item.Roles))
			for _, item := range item.Roles {
				items = append(items, StaffRole{ID: item.ID, Position: item.Position, Name: item.Name, Color: item.Color})
			}
			return items
		}()})
	}
	// Get user access
	access := res.Access

	// Return data
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access": access,
		"staff":  sMembers,
	})
}

func (h *Handler) DeleteStaffMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	targetStaffID, err := strconv.ParseInt(chi.URLParam(r, "staffID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.App.Staff.Members.DeleteStaffMember(ctx, &members.DeleteStaffMemberInput{
		UserID:  userID,
		StaffID: targetStaffID,
	})
	if err != nil {
		switch {
		case errors.Is(err, members.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, members.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) LeaveSchool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Members.LeaveSchoolStaff(ctx, &members.LeaveSchoolStaffInput{
		UserID:   userID,
		SchoolID: schoolID,
	})
	if err != nil {
		switch {
		case errors.Is(err, members.ErrForbidden):
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
