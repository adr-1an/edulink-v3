package staff

import (
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func ListSchoolLogsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !schools.Can(schools.PermissionLogsList, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

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
	var logs []schoolLog

	rows, err := db.QueryContext(ctx, `
		SELECT
		    l.id,
		    l.action,
		    l.type,
		    l.title,
		    l.message,
		    l.details,
		    l.created_at,
		    u.id,
		    u.name,
		    u.email
		FROM school_logs l
		JOIN users u ON u.id = l.by_user
		WHERE school_id = $1
		AND l.created_at > NOW() - INTERVAL '30 days'
	`, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var l schoolLog

		if err := rows.Scan(
			&l.ID,
			&l.Action,
			&l.Type,
			&l.Title,
			&l.Message,
			&l.Details,
			&l.CreatedAt,
			&l.User.ID,
			&l.User.Name,
			&l.User.Email,
		); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logs = append(logs, l)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	access, err := schools.GetAllUserPermissions(ctx, db, userID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"access": access,
		"logs":   logs,
	})
}
