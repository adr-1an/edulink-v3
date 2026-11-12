package portal

import (
	"app/internal/application/portal/auth"
	"app/internal/helpers"
	"app/internal/helpers/portal"
	"app/internal/helpers/staff"

	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type payload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var p payload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Portal.Auth.Login(ctx, &auth.LoginInput{
		Email:    p.Email,
		Password: p.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUnauthorized):
			w.WriteHeader(http.StatusUnauthorized)
		case errors.Is(err, auth.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"token": res.Token,
	})
}

func (h *Handler) TokenCheckHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := portal.TokenToUID(w, r, h.DB, ctx, helpers.AccTypeEither)
	if err != nil {
		return
	}

	res, err := h.App.Portal.Auth.TokenCheck(ctx, &auth.TokenCheckInput{UserID: userID})
	if err != nil {
		if errors.Is(err, auth.ErrUnauthorized) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"user": map[string]any{
			"id":          userID,
			"accountType": res.AccountType,
		},
	})
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token, err := portal.ParseToken(w, r)
	if err != nil {
		return
	}
	if err := h.App.Portal.Auth.Logout(ctx, &auth.LogoutInput{Token: token}); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AccountActivationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	activationToken := chi.URLParam(r, "token")
	if activationToken == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staffhelpers.ErrorCodeNoToken,
		})
		return
	}
	type payload struct {
		NewPassword string `json:"newPassword"`
	}
	var p payload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)
	res, err := h.App.Portal.Auth.Activate(ctx, &auth.ActivateInput{Token: activationToken, NewPassword: p.NewPassword, InvalidPayload: decodeErr != nil, Session: auth.SessionInput{ForwardedFor: r.Header.Get("X-Forwarded-For"), SourceIP: r.RemoteAddr, UserAgent: r.Header.Get("User-Agent")}})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidToken):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": staffhelpers.ErrorCodeInvalidToken})
		case errors.Is(err, auth.ErrExpiredToken):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": staffhelpers.ErrorCodeExpiredToken})
		case errors.Is(err, auth.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, auth.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"token": res.Token})
}

func (h *Handler) PasswordChangeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := portal.TokenToUID(w, r, h.DB, ctx, helpers.AccTypeEither)
	if err != nil {
		return
	}

	type payload struct {
		Current string `json:"currentPassword"`
		New     string `json:"newPassword"`
	}
	var p payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.App.Portal.Auth.UpdatePassword(ctx, &auth.UpdatePasswordInput{
		UserID:      userID,
		OldPassword: p.Current,
		NewPassword: p.New,
	})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUnauthorized):
			w.WriteHeader(http.StatusUnauthorized)

		case errors.Is(err, auth.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
