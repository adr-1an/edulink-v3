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
	"app/internal/responsecodes"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListUserInvitations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	type User struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	type School struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		RegionCode string `json:"regionCode"`
	}
	type Invitation struct {
		ID        string    `json:"id"`
		School    School    `json:"school"`
		SentBy    User      `json:"sentBy"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"createdAt"`
		ExpiresAt time.Time `json:"expiresAt"`
	}
	res, err := h.App.Staff.Members.ListUserInvitations(ctx, &members.ListUserInvitationsInput{UserID: userID})
	if err != nil {
		switch {
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	var invitations []Invitation
	if res.Invitations != nil {
		invitations = make([]Invitation, 0, len(res.Invitations))
	}
	for _, item := range res.Invitations {
		invitations = append(invitations, Invitation{ID: item.ID, School: School{ID: item.School.ID, Name: item.School.Name, RegionCode: item.School.RegionCode}, SentBy: User{ID: item.SentBy.ID, Name: item.SentBy.Name, Email: item.SentBy.Email}, Status: item.Status, CreatedAt: item.CreatedAt, ExpiresAt: item.ExpiresAt})
	}
	// Return invitations
	_ = json.NewEncoder(w).Encode(map[string]any{
		"invitations": invitations,
	})
}

func (h *Handler) ListSchoolInvitations(w http.ResponseWriter, r *http.Request) {
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

	type user struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	type invite struct {
		ID        string    `json:"id"`
		UserEmail string    `json:"userEmail"`
		SentBy    user      `json:"addedBy"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"createdAt"`
		ExpiresAt time.Time `json:"expiresAt"`
	}
	res, err := h.App.Staff.Members.ListSchoolInvitations(ctx, &members.ListSchoolInvitationsInput{UserID: userID, SchoolID: schoolID})
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
	var invites []invite
	if res.Invitations != nil {
		invites = make([]invite, 0, len(res.Invitations))
	}
	for _, item := range res.Invitations {
		invites = append(invites, invite{ID: item.ID, UserEmail: item.UserEmail, SentBy: user{ID: item.SentBy.ID, Name: item.SentBy.Name, Email: item.SentBy.Email}, Status: item.Status, CreatedAt: item.CreatedAt, ExpiresAt: item.ExpiresAt})
	}
	access := res.Access

	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":      access,
		"invitations": invites,
	})
}

func (h *Handler) CancelInvitation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	invID, err := strconv.ParseInt(chi.URLParam(r, "invitationID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Members.CancelInvitation(ctx, &members.CancelInvitationInput{
		UserID:       userID,
		InvitationID: invID,
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

func (h *Handler) ViewInvitation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// No auth required for this route, token is enough to view

	// Get token from URL & hash it
	token := chi.URLParam(r, "token")
	type StaffInvitation struct {
		ID          string    `json:"id"`
		SentToEmail string    `json:"sentToEmail"`
		SchoolName  string    `json:"schoolName"`
		SentByName  string    `json:"sentByName"`
		SentByEmail string    `json:"sentByEmail"`
		Status      string    `json:"status"`
		CreatedAt   time.Time `json:"createdAt"`
		ExpiresAt   time.Time `json:"expiresAt"`
	}
	res, err := h.App.Staff.Members.ViewInvitation(ctx, &members.ViewInvitationInput{Token: token})
	if err != nil {
		switch {
		case errors.Is(err, members.ErrInvalidToken):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeInvalidToken})
		case errors.Is(err, members.ErrExpired):
			w.WriteHeader(http.StatusGone)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	inv := StaffInvitation{ID: res.Value.ID, SentToEmail: res.Value.SentToEmail, SchoolName: res.Value.SchoolName, SentByName: res.Value.SentByName, SentByEmail: res.Value.SentByEmail, Status: res.Value.Status, CreatedAt: res.Value.CreatedAt, ExpiresAt: res.Value.ExpiresAt}
	// Return invitation
	_ = json.NewEncoder(w).Encode(map[string]any{
		"invitation": inv,
	})
}

func (h *Handler) SendStaffInvitation(w http.ResponseWriter, r *http.Request) {
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

	// Payload
	type Payload struct {
		Email      string `json:"email"`
		Importance string `json:"importance"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)

	res, err := h.App.Staff.Members.SendInvitation(ctx, &members.SendInvitationInput{
		InvalidPayload: decodeErr != nil,
		UserID:         userID,
		SchoolID:       schoolID,
		Invitation:     members.SendInvitation{Email: p.Email, Importance: p.Importance},
	})
	if err != nil {
		switch {
		case errors.Is(err, members.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, members.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, members.ErrInvalidEmail):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeInvalidEmail})
		case errors.Is(err, members.ErrInvalidMailImportance):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeInvalidMailImportance})
		case errors.Is(err, members.ErrStaffMemberEmailConflict):
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeStaffMemberEmailConflict})
		case errors.Is(err, members.ErrInvitationEmailConflict):
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeInvitationEmailConflict})
		case errors.Is(err, members.ErrTargetPrivacyRestricted):
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeTargetPrivacyRestricted})
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if res.NoRowsAffected {
		w.WriteHeader(http.StatusInternalServerError)
	}
	if res.StaffCreated {
		w.WriteHeader(http.StatusCreated)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AcceptStaffInvitation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get token from URL & hash it
	token := chi.URLParam(r, "token")

	res, err := h.App.Staff.Members.AcceptInvitation(ctx, &members.AcceptInvitationInput{
		UserID: userID,
		Token:  token,
	})
	if err != nil {
		switch {
		case errors.Is(err, members.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, members.ErrInvalidToken):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeInvalidToken})
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

func (h *Handler) RejectStaffInvitation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get invitation token & hash it
	token := chi.URLParam(r, "token")

	res, err := h.App.Staff.Members.RejectInvitation(ctx, &members.RejectInvitationInput{
		UserID: userID,
		Token:  token,
	})
	if err != nil {
		switch {
		case errors.Is(err, members.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, members.ErrInvalidToken):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeInvalidToken})
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

func (h *Handler) AcceptStaffInvitationByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get invitation ID
	invitationIDStr := chi.URLParam(r, "invitationID")
	invitationID, err := strconv.ParseInt(invitationIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Members.AcceptInvitationByID(ctx, &members.AcceptInvitationByIDInput{
		UserID:       userID,
		InvitationID: invitationID,
	})
	if err != nil {
		switch {
		case errors.Is(err, members.ErrInvalidToken):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeInvalidToken})
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

func (h *Handler) RejectStaffInvitationByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get invitation ID
	invitationIDStr := chi.URLParam(r, "invitationID")
	invitationID, err := strconv.ParseInt(invitationIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Members.RejectInvitationByID(ctx, &members.RejectInvitationByIDInput{
		UserID:       userID,
		InvitationID: invitationID,
	})
	if err != nil {
		switch {
		case errors.Is(err, members.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, members.ErrInvalidToken):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeInvalidToken})
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
