package staff

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"app/internal/application/staff/auth"
	staffhelpers "app/internal/helpers/staff"

	"github.com/go-chi/chi/v5"
)

// SendRegistrationLinkHandler generates a registration link, and if the provided email isn't taken,
// sends a registration email.
func (h *Handler) SendRegistrationLinkHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Payload
	type Payload struct {
		Email string `json:"email"`
	}
	var p Payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := h.App.Staff.Auth.SendRegistrationLink(ctx, &auth.SendRegistrationLinkInput{
		Email: p.Email,
	})

	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)

	case errors.Is(err, auth.ErrUnprocessableEntity):
		w.WriteHeader(http.StatusUnprocessableEntity)

	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// CheckRegistrationTokenHandler checks whether the given registration token is valid
// and returns the email associated with it.
func (h *Handler) CheckRegistrationTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := chi.URLParam(r, "token")

	res, err := h.App.Staff.Auth.CheckRegistrationToken(ctx, &auth.CheckRegistrationTokenInput{
		Token: token,
	})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"email": res.Email,
	})
}

// RegistrationHandler - Main handler for user registration.
// Validation rules:
// Name - min 1, max 128
// Phone - min 3, max 32
// Password - min 8, no max
func (h *Handler) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Payload
	type Payload struct {
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token := chi.URLParam(r, "token")

	err := h.App.Staff.Auth.Register(ctx, &auth.RegisterInput{
		Name:              p.Name,
		Phone:             p.Phone,
		Password:          p.Password,
		RegistrationToken: token,
	})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)

		case errors.Is(err, auth.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	w.WriteHeader(http.StatusCreated)
}

// LoginHandler - Main handler for user login.
// Validation rules:
// Email - min 5, max 254
// Password - min 8, no max
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user agent & IP
	userAgent := r.Header.Get("User-Agent")
	forwardedFor := r.Header.Get("X-Forwarded-For")
	var ip *string
	if forwardedFor != "" {
		ip = &forwardedFor
	}

	// Payload
	type Payload struct {
		Email        string `json:"email"`
		Password     string `json:"password"`
		StayLoggedIn bool   `json:"stayLoggedIn"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Auth.Login(ctx, &auth.LoginInput{
		Email:        p.Email,
		Password:     p.Password,
		StayLoggedIn: p.StayLoggedIn,
		IP:           ip,
		UserAgent:    userAgent,
	})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)

		case errors.Is(err, auth.ErrUnauthorized):
			w.WriteHeader(http.StatusUnauthorized)
		}

		return
	}

	switch res.Type {
	case auth.LoginResultSuccess:
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token": res.Token,
		})

	case auth.LoginResultTwoFactor:
		_ = json.NewEncoder(w).Encode(map[string]any{
			"twoFactorChallenge": map[string]any{
				"token":     res.TwoFactorChallenge.Token,
				"purpose":   res.TwoFactorChallenge.Purpose,
				"expiresAt": res.TwoFactorChallenge.ExpiresAt,
			},
		})
	}
}

// TokenCheckHandler checks whether the given token is valid and not expired.
// It will only return a 204 if the token is valid.
func (h *Handler) TokenCheckHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get token & hash
	token := r.Header.Get("Authorization")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staffhelpers.ErrorCodeNoToken,
		})
		return
	}
	token = strings.TrimPrefix(token, "Bearer ")

	valid, err := h.App.Staff.Auth.CheckToken(ctx, &auth.CheckTokenInput{
		Token: token,
	})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUnauthorized):
			w.WriteHeader(http.StatusUnauthorized)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SendPasswordResetEmailHandler sends a password reset link ("forgot password") to the user's email address.
func (h *Handler) SendPasswordResetEmailHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Payload
	type Payload struct {
		Email string `json:"email"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := h.App.Staff.Auth.SendPasswordResetEmail(ctx, &auth.SendPasswordResetEmailInput{Email: p.Email})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": staffhelpers.ErrorCodeInvalidEmail})
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) PasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get reset token
	token := chi.URLParam(r, "token")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staffhelpers.ErrorCodeNoToken,
		})
		return
	}

	// Payload
	type Payload struct {
		NewPassword string `json:"newPassword"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := h.App.Staff.Auth.ResetPassword(ctx, &auth.ResetPasswordInput{Token: token, NewPassword: p.NewPassword})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, auth.ErrUnauthorized):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": staffhelpers.ErrorCodeInvalidToken})
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tokenHash, err := staffhelpers.TokenToHash(w, r)
	if err != nil {
		return
	}
	err = h.App.Staff.Auth.Logout(ctx, &auth.LogoutInput{TokenHash: tokenHash})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUnauthorized):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": staffhelpers.ErrorCodeInvalidToken})
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Enable2faHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}
	res, err := h.App.Staff.Auth.EnableTwoFactor(ctx, &auth.EnableTwoFactorInput{UserID: userID})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, auth.ErrConflict):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"url": res.URL})
}

func (h *Handler) Verify2faHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	type payload struct {
		TOTP string `json:"code"`
	}
	var p payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res, err := h.App.Staff.Auth.VerifyTwoFactor(ctx, &auth.VerifyTwoFactorInput{UserID: userID, TOTP: p.TOTP})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, auth.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, auth.ErrUnauthorized):
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"codes": res.Codes})
}

func (h *Handler) Disable2faHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	type payload struct {
		TOTP     string `json:"code"`
		Password string `json:"password"`
	}
	var p payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.App.Staff.Auth.DisableTwoFactor(ctx, &auth.DisableTwoFactorInput{UserID: userID, TOTP: p.TOTP, Password: p.Password})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, auth.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, auth.ErrUnauthorized):
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RecoverTwoFactorHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	type payload struct {
		RecoveryCode string `json:"recoveryCode"`
	}
	var p payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := h.App.Staff.Auth.RecoverTwoFactor(ctx, &auth.RecoverTwoFactorInput{RecoveryCode: p.RecoveryCode})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": staffhelpers.ErrorCodeInvalidRecoveryCode})
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
