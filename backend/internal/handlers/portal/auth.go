package portal

import (
	authService "app/internal/application/portal/auth"
	"app/internal/auth"
	"app/internal/portalusers"
	"app/internal/responsecodes"

	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type payload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var p payload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Portal.Auth.Login(ctx, &authService.LoginInput{
		Email:    p.Email,
		Password: p.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, authService.ErrUnauthorized):
			w.WriteHeader(http.StatusUnauthorized)
		case errors.Is(err, authService.ErrUnprocessableEntity):
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

func (h *Handler) CheckToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUID(w, r, h.DB, ctx, portalusers.AccTypeEither)
	if err != nil {
		return
	}

	res, err := h.App.Portal.Auth.TokenCheck(ctx, &authService.TokenCheckInput{UserID: userID})
	if err != nil {
		if errors.Is(err, authService.ErrUnauthorized) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"user": map[string]any{
			"id":          strconv.FormatInt(userID, 10),
			"accountType": res.AccountType,
		},
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token, err := auth.ParseToken(w, r)
	if err != nil {
		return
	}
	if err := h.App.Portal.Auth.Logout(ctx, &authService.LogoutInput{Token: token}); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ActivateAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	activationToken := chi.URLParam(r, "token")
	if activationToken == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": responsecodes.ErrorCodeNoToken,
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
	res, err := h.App.Portal.Auth.Activate(ctx, &authService.ActivateInput{Token: activationToken, NewPassword: p.NewPassword, InvalidPayload: decodeErr != nil, Session: authService.SessionInput{ForwardedFor: r.Header.Get("X-Forwarded-For"), SourceIP: r.RemoteAddr, UserAgent: r.Header.Get("User-Agent")}})
	if err != nil {
		switch {
		case errors.Is(err, authService.ErrInvalidToken):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeInvalidToken})
		case errors.Is(err, authService.ErrExpiredToken):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeExpiredToken})
		case errors.Is(err, authService.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, authService.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"token": res.Token})
}

func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUID(w, r, h.DB, ctx, portalusers.AccTypeEither)
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

	err = h.App.Portal.Auth.UpdatePassword(ctx, &authService.UpdatePasswordInput{
		UserID:      userID,
		OldPassword: p.Current,
		NewPassword: p.New,
	})
	if err != nil {
		switch {
		case errors.Is(err, authService.ErrUnauthorized):
			w.WriteHeader(http.StatusUnauthorized)

		case errors.Is(err, authService.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
