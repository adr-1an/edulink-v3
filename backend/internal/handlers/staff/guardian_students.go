package staff

import (
	"app/internal/application/staff/guardians"
	"app/internal/auth"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListGuardianStudents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
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

	type student struct {
		ID          string    `json:"id"`
		Name        string    `json:"name"`
		LastName    string    `json:"lastName"`
		DateOfBirth time.Time `json:"dateOfBirth"`
	}
	students := make([]student, 0, len(res.Students))
	for _, s := range res.Students {
		students = append(students, student{
			ID:          strconv.FormatInt(s.ID, 10),
			Name:        s.Name,
			LastName:    s.LastName,
			DateOfBirth: s.DoB,
		})
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":   res.Access,
		"students": students,
	})
}
