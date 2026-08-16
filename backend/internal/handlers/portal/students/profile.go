package students

import (
	"app/internal/helpers"
	"app/internal/helpers/portal"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

func GetProfileHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := portal.TokenToUID(w, r, db, ctx, helpers.AccTypeEither)
	if err != nil {
		return
	}

	type user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type school struct {
		Owner  user   `json:"owner"`
		Name   string `json:"name"`
		Region string `json:"region"`
	}

	type profile struct {
		Name        string `json:"name"`
		LastName    string `json:"lastName"`
		Email       string `json:"email"`
		Phone       string `json:"phone"`
		DateOfBirth string `json:"dateOfBirth"`
		AccountType string `json:"accountType"`

		School school `json:"school"`
	}
	var p profile

	if err := db.QueryRowContext(ctx, `
		SELECT
		    u.name, u.email,
		    s.name, s.region_code,
		    pu.name, pu.last_name, pu.email, pu.date_of_birth, COALESCE(pu.phone, ''), account_type
		FROM portal_users pu
		JOIN schools s
		ON pu.school_id = s.id
		JOIN users u
		ON s.owner_id = u.id
		WHERE pu.id = $1
	`, userID).
		Scan(
			&p.School.Owner.Name,
			&p.School.Owner.Email,

			&p.School.Name,
			&p.School.Region,

			&p.Name,
			&p.LastName,
			&p.Email,
			&p.DateOfBirth,
			&p.Phone,
			&p.AccountType,
		); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"profile": p,
	})
}
