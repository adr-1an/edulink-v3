package staff

import (
	"errors"
	"net/http"
	"strconv"

	"app/internal/application/staff/portalusers"
	"app/internal/auth"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) SendPortalUserAccountVerificationEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	portalUserID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.App.Staff.PortalUsers.SendActivationEmail(ctx, &portalusers.SendActivationEmailInput{
		UserID:       userID,
		PortalUserID: portalUserID,
	}); err != nil {
		switch {
		case errors.Is(err, portalusers.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, portalusers.ErrAccountDisabled):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, portalusers.ErrAccountAlreadyActive):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
