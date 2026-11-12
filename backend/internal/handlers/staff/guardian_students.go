package staff

import (
	"app/internal/application/staff/guardians"
	"app/internal/helpers/staff"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListGuardianStudentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	guardianID, err := strconv.ParseInt(chi.URLParam(r, "guardianID"), 10, 64)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Guardians.ListGuardianStudents(ctx, &guardians.ListGuardianStudentsInput{
		UserID:     userID,
		GuardianID: guardianID,
	})
	if err != nil {
		switch {
		case errors.Is(err, guardians.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	access := res.Access

	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":   access,
		"students": res.Students,
	})
}
