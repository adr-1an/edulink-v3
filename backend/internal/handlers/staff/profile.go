package staff

import (
	"encoding/json"
	"errors"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"strconv"

	staffProfile "app/internal/application/staff/profile"
	"app/internal/auth"
	"app/internal/responsecodes"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}
	res, err := h.App.Staff.Profile.GetProfile(ctx, &staffProfile.GetProfileInput{UserID: userID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var picture any
	if res.Pfp != nil {
		picture = map[string]any{"presignedUrl": res.Pfp.PresignedURL, "fileName": res.Pfp.Filename, "contentType": res.Pfp.ContentType}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"user": map[string]any{
		"profilePicture": picture, "id": strconv.FormatInt(res.ID, 10), "name": res.Name, "email": res.Email, "phone": res.Phone,
		"twoFactorStatus": res.TwoFactorStatus, "publicProfile": res.PublicProfile, "staffInvitationsDisabled": res.InvDisabled, "updatedAt": res.UpdatedAt, "createdAt": res.CreatedAt,
	}})
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Payload
	type Payload struct {
		Name          string `json:"name"`
		Phone         string `json:"phone"`
		PublicProfile bool   `json:"publicProfile"`
		InvDisabled   bool   `json:"staffInvitationsDisabled"`
	}
	var p Payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.App.Staff.Profile.UpdateProfile(ctx, &staffProfile.UpdateProfileInput{UserID: userID, Name: p.Name, Phone: p.Phone, PublicProfile: p.PublicProfile, InvDisabled: p.InvDisabled})
	if err != nil {
		switch {
		case errors.Is(err, staffProfile.ErrInvalidData):
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SendEmailChange(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Payload
	type Payload struct {
		NewEmail string `json:"newEmail"`
		Password string `json:"password"`
	}
	var p Payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.App.Staff.Profile.SendEmailChange(ctx, &staffProfile.SendEmailChangeInput{UserID: userID, NewEmail: p.NewEmail, Password: p.Password})
	if err != nil {
		switch {
		case errors.Is(err, staffProfile.ErrInvalidData):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, staffProfile.ErrIncorrectPassword):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeIncorrectPassword})
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get token from URL
	token := chi.URLParam(r, "token")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := h.App.Staff.Profile.UpdateEmail(ctx, &staffProfile.UpdateEmailInput{Token: token})
	if err != nil {
		switch {
		case errors.Is(err, staffProfile.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, staffProfile.ErrConflict):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Payload
	type Payload struct {
		Password    string `json:"password"`
		NewPassword string `json:"newPassword"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.App.Staff.Profile.ChangePassword(ctx, &staffProfile.ChangePasswordInput{UserID: userID, Password: p.Password, NewPassword: p.NewPassword})
	if err != nil {
		switch {
		case errors.Is(err, staffProfile.ErrInvalidData):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, staffProfile.ErrUnauthorized):
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UploadProfilePicture(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	type payload struct {
		Name                string `json:"name"`
		DeclaredSize        int64  `json:"declaredSize"`
		DeclaredContentType string `json:"declaredContentType"`
	}
	var p payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res, err := h.App.Staff.Profile.UploadProfilePicture(ctx, &staffProfile.UploadProfilePictureInput{UserID: userID, Name: p.Name, DeclaredSize: p.DeclaredSize, DeclaredContentType: p.DeclaredContentType})
	if err != nil {
		switch {
		case errors.Is(err, staffProfile.ErrInvalidData):
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"id": strconv.FormatInt(res.ID, 10), "completionToken": res.CompletionToken, "url": res.URL})
}

func (h *Handler) ClearProfilePicture(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}
	err = h.App.Staff.Profile.ClearProfilePicture(ctx, &staffProfile.ClearProfilePictureInput{UserID: userID})
	if err != nil {
		switch {
		case errors.Is(err, staffProfile.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
