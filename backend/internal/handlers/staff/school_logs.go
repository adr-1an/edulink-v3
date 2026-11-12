package staff

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"app/internal/application/staff/logs"
	"app/internal/auth"

	"github.com/go-chi/chi/v5"
)

type user struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type schoolLog struct {
	ID        string `json:"id"`
	User      user   `json:"user"`
	Action    string `json:"action"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Details   string `json:"details"`
	CreatedAt string `json:"createdAt"`
}

func (h *Handler) ListSchoolLogsHandler(w http.ResponseWriter, r *http.Request) {
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

	res, err := h.App.Staff.Logs.ListLogs(ctx, &logs.ListLogsInput{
		UserID:   userID,
		SchoolID: schoolID,
	})
	if err != nil {
		switch {
		case errors.Is(err, logs.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	logs := make([]schoolLog, 0)
	for _, l := range res.Logs {
		logs = append(logs, schoolLog{
			ID: strconv.FormatInt(l.ID, 10),
			User: user{
				ID:    strconv.FormatInt(l.User.ID, 10),
				Name:  l.User.Name,
				Email: l.User.Email,
			},
			Action:    l.Action,
			Type:      l.Type,
			Title:     l.Title,
			Message:   l.Message,
			Details:   l.Details,
			CreatedAt: l.CreatedAt,
		})
	}

	access := res.Access

	_ = json.NewEncoder(w).Encode(map[string]any{
		"access": access,
		"logs":   logs,
	})
}
