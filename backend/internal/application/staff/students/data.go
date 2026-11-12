package students

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"
)

type StudentPayload struct {
	Name     string `json:"name"`
	LastName string `json:"lastName"`
	DoB      string `json:"dob"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Notes    string `json:"notes"`

	AccountEnabled bool   `json:"accountEnabled"`
	AccountActive  bool   `json:"accountActive"`
	Password       string `json:"password"`
}

func parsePayload(p StudentPayload) (StudentPayload, error) {
	p.Name = strings.TrimSpace(p.Name)
	p.LastName = strings.TrimSpace(p.LastName)
	p.DoB = strings.TrimSpace(p.DoB)
	if _, err := time.Parse("2006-01-02", p.DoB); err != nil {
		return StudentPayload{}, errors.New("invalid date of birth")
	}

	p.Email = strings.TrimSpace(strings.ToLower(p.Email))
	p.Phone = strings.TrimSpace(p.Phone)
	p.Notes = strings.TrimSpace(p.Notes)

	return p, nil
}

func validate(p StudentPayload) (StudentPayload, error) {
	p, err := parsePayload(p)
	if err != nil {
		return p, err
	}

	if p.Name == "" || len(p.Name) < 1 || len(p.Name) > 32 ||
		p.LastName == "" || len(p.LastName) < 1 || len(p.LastName) > 32 ||
		p.Email == "" || len(p.Email) < 5 || len(p.Email) > 254 ||
		p.Phone != "" && len(p.Phone) < 3 || len(p.Phone) > 32 ||
		p.Notes != "" && len(p.Notes) > 2048 {
		return StudentPayload{}, errors.New("validation failed")
	}

	return p, nil
}

func checkSchoolAndEmailConflict(tx *sql.Tx, ctx context.Context, email string) (bool, error) {
	var conflict bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM portal_users WHERE email = $1
		)
	`, email).Scan(&conflict); err != nil {
		log.Println(err)
		return false, err
	}

	return conflict, nil
}
