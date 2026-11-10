package staff

import (
	staffStudents "app/internal/application/staff/students"
	"app/internal/helpers"
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/go-chi/chi/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/wneessen/go-mail"
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

func (h *Handler) CreateStudentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !schools.Can(schools.PermissionStudentCreate, userID, schoolID, ctx, h.DB) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var p StudentPayload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err = validate(p)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if p.AccountActive && !schools.Can(schools.PermissionPortalUserActivate, userID, schoolID, ctx, h.DB) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	conflict, err := checkSchoolAndEmailConflict(tx, ctx, p.Email)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if conflict {
		w.WriteHeader(http.StatusConflict)
		return
	}

	id, err := h.Sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if (p.AccountEnabled && p.Password == "") ||
		(p.Password != "" && len(p.Password) < 8) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	var passHash *string
	if p.Password != "" {
		hash, err := argon2id.CreateHash(p.Password, argon2id.DefaultParams)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		passHash = &hash
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO portal_users (
			id,
			school_id,
			name,
			last_name,
			date_of_birth,
			email,
			phone,
			notes,
			account_enabled,
		  	account_active,
			password_hash,
			account_type
			)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`,
		id,
		schoolID,
		p.Name,
		p.LastName,
		p.DoB,
		p.Email,
		p.Phone,
		p.Notes,
		p.AccountEnabled,
		p.AccountActive,
		passHash,
		helpers.AccTypeStudent,
	)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	details := fmt.Sprintf("{user} created a new student: %s <%s>", p.Name, p.Email)
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionStudentCreate, schools.TypeCreate, "Student created", "{user} created a student.", tx, ctx, h.Sf, details); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staffhelpers.RowsAffectedOr500(res, w); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) UpdateStudentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	studentID, err := strconv.ParseInt(chi.URLParam(r, "studentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	schoolID, err := staffhelpers.StudentToSchoolID(ctx, h.DB, studentID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !schools.Can(schools.PermissionStudentUpdate, userID, schoolID, ctx, h.DB) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var p StudentPayload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err = validate(p)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	conflict, err := checkSchoolAndEmailConflict(tx, ctx, p.Email)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if conflict {
		w.WriteHeader(http.StatusConflict)
		return
	}

	var accountActive bool
	if err := tx.QueryRowContext(ctx, `
		SELECT account_active
		FROM portal_users
		WHERE id = $1
		AND school_id = $2
		AND account_type = $3
	`, studentID, schoolID, helpers.AccTypeStudent).Scan(&accountActive); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if permission, changed := schools.PortalUserActiveStatePermission(accountActive, p.AccountActive); changed &&
		!schools.Can(permission, userID, schoolID, ctx, h.DB) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if p.Password != "" && len(p.Password) < 8 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	var passHash *string
	if p.Password != "" {
		hash, err := argon2id.CreateHash(p.Password, argon2id.DefaultParams)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		passHash = &hash
	}

	var (
		name     string
		lastName string
		email    string
	)
	if err := h.DB.QueryRowContext(ctx, `
		SELECT name, last_name, email
		FROM portal_users
		WHERE id = $1
		AND school_id = $2
		AND account_type = $3
	`, studentID, schoolID, helpers.AccTypeStudent).Scan(&name, &lastName, &email); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE portal_users
		SET
		    name = $1,
		    last_name = $2,
		    date_of_birth = $3,
		    email = $4,
		    phone = $5,
		    notes = $6,
		    account_enabled = $7,
		    account_active = $8,
		    password_hash = COALESCE($9, password_hash)
		WHERE id = $10
		AND school_id = $11
		AND account_type = $12
	`,
		p.Name,
		p.LastName,
		p.DoB,
		p.Email,
		p.Phone,
		p.Notes,
		p.AccountEnabled,
		p.AccountActive,
		passHash,
		studentID,
		schoolID,
		helpers.AccTypeStudent,
	)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if passHash != nil {
		// Password has been changed, delete all sessions
		if _, err := h.DB.ExecContext(ctx, `
			DELETE FROM portal_sessions
			WHERE portal_user_id = $1
		`, studentID); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	details := fmt.Sprintf(
		"{user} updated the student %s %s <%s> to %s %s <%s>.",
		name, lastName, email, p.Name, p.LastName, p.Email)
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionStudentEdit, schools.TypeEdit, "Student updated", "{user} updated a student.", tx, ctx, h.Sf, details); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staffhelpers.RowsAffectedOr500(res, w); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteStudentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	studentID, err := strconv.ParseInt(chi.URLParam(r, "studentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	schoolID, err := staffhelpers.StudentToSchoolID(ctx, h.DB, studentID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !schools.Can(schools.PermissionStudentDelete, userID, schoolID, ctx, h.DB) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	var (
		name  string
		email string
	)
	if err := tx.QueryRowContext(ctx, `
		DELETE FROM portal_users
		WHERE id = $1
		AND school_id = $2
		AND account_type = $3
		RETURNING name, email
	`, studentID, schoolID, helpers.AccTypeStudent).Scan(&name, &email); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	details := fmt.Sprintf("{user} deleted the student %s <%s>.", name, email)
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionStudentDelete, schools.TypeDelete, "Student deleted", "{user} deleted a student.", tx, ctx, h.Sf, details); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListStudentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	type student struct {
		ID                string     `json:"id"`
		Name              string     `json:"name"`
		LastName          string     `json:"lastName"`
		ProfilePictureURL *string    `json:"profilePictureURL"`
		DoB               *time.Time `json:"dateOfBirth"`
		Email             *string    `json:"email"`
		Phone             *string    `json:"phone"`
		Notes             *string    `json:"notes"`
		AccountEnabled    bool       `json:"accountEnabled"`
		AccountActive     bool       `json:"accountActive"`
		CreatedAt         time.Time  `json:"createdAt"`
	}
	var studentList []student

	res, err := h.App.Staff.Students.ListStudents(ctx, &staffStudents.ListStudentsInput{
		UserID:   userID,
		SchoolID: schoolID,
	})
	if err != nil {
		switch {
		case errors.Is(err, staffStudents.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	if res != nil {
		for _, s := range res.Students {
			idStr := strconv.FormatInt(s.ID, 10)

			studentList = append(studentList, student{
				ID:                idStr,
				Name:              s.Name,
				LastName:          s.LastName,
				ProfilePictureURL: s.ProfilePictureURL,
				DoB:               s.DoB,
				Email:             s.Email,
				Phone:             s.Phone,
				Notes:             s.Notes,
				AccountEnabled:    s.AccountEnabled,
				AccountActive:     s.AccountActive,
				CreatedAt:         s.CreatedAt,
			})
		}
	}

	access, err := schools.GetAllUserPermissions(ctx, h.DB, userID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":   access,
		"students": studentList,
	})
}

func (h *Handler) ViewStudentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	studentID, err := strconv.ParseInt(chi.URLParam(r, "studentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Students.ViewStudent(ctx, &staffStudents.ViewStudentInput{
		UserID:    userID,
		StudentID: studentID,
	})
	if err != nil {
		switch {
		case errors.Is(err, staffStudents.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	access, err := schools.GetAllUserPermissions(ctx, h.DB, userID, res.SchoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload := studentViewPayload(res)
	payload["access"] = access
	_ = json.NewEncoder(w).Encode(payload)
}

func studentViewPayload(res *staffStudents.ViewStudentResult) map[string]any {
	type grade struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Level int    `json:"level"`
	}

	type course struct {
		ID    string `json:"id"`
		Grade grade  `json:"grade"`
		Name  string `json:"name"`
		Color string `json:"accentColor"`
	}

	type assignment struct {
		ID          string  `json:"id"`
		Course      course  `json:"course"`
		Title       string  `json:"title"`
		Description *string `json:"description"`
	}

	type assignmentScore struct {
		ScorePercentage int `json:"scorePercentage"`
	}

	type submission struct {
		ID          string           `json:"id"`
		Assignment  assignment       `json:"assignment"`
		Score       *assignmentScore `json:"score"`
		SubmittedAt time.Time        `json:"submittedAt"`
		Notes       *string          `json:"notes"`
	}
	var submissions []submission

	type profilePicture struct {
		PresignedURL string `json:"presignedUrl"`
	}

	type guardian struct {
		ID       string  `json:"id"`
		Name     string  `json:"name"`
		LastName string  `json:"lastName"`
		Email    string  `json:"email"`
		Phone    *string `json:"phone"`
	}

	type student struct {
		ProfilePicture *profilePicture `json:"profilePicture"`
		ID             string          `json:"id"`
		Name           string          `json:"name"`
		LastName       string          `json:"lastName"`
		DoB            string          `json:"dateOfBirth"`
		Email          string          `json:"email"`
		Phone          *string         `json:"phone"`
		Notes          *string         `json:"notes"`
		AccountEnabled bool            `json:"accountEnabled"`
		AccountActive  bool            `json:"accountActive"`
		CreatedAt      time.Time       `json:"createdAt"`
		Guardians      []guardian      `json:"guardians"`
	}

	for _, s := range res.Submissions {
		sub := submission{
			ID: strconv.FormatInt(s.ID, 10),
			Assignment: assignment{
				ID: strconv.FormatInt(s.Assignment.ID, 10),
				Course: course{
					ID: strconv.FormatInt(s.Assignment.Course.ID, 10),
					Grade: grade{
						ID:    strconv.FormatInt(s.Assignment.Course.Grade.ID, 10),
						Name:  s.Assignment.Course.Grade.Name,
						Level: s.Assignment.Course.Grade.Level,
					},
					Name:  s.Assignment.Course.Name,
					Color: s.Assignment.Course.Color,
				},
				Title:       s.Assignment.Title,
				Description: s.Assignment.Description,
			},
			Score:       nil,
			SubmittedAt: s.SubmittedAt,
			Notes:       s.Notes,
		}

		if s.Score != nil {
			sub.Score = &assignmentScore{
				ScorePercentage: s.Score.ScorePercentage,
			}
		}

		submissions = append(submissions, sub)
	}

	var guardians []guardian
	for _, g := range res.Student.Guardians {
		guardians = append(guardians, guardian{
			ID:       strconv.FormatInt(g.ID, 10),
			Name:     g.Name,
			LastName: g.LastName,
			Email:    g.Email,
			Phone:    g.Phone,
		})
	}
	s := student{
		ID:             strconv.FormatInt(res.Student.ID, 10),
		Name:           res.Student.Name,
		LastName:       res.Student.LastName,
		DoB:            res.Student.DoB,
		Email:          res.Student.Email,
		Phone:          res.Student.Phone,
		Notes:          res.Student.Notes,
		AccountEnabled: res.Student.AccountEnabled,
		AccountActive:  res.Student.AccountActive,
		CreatedAt:      res.Student.CreatedAt,
		Guardians:      guardians,
	}

	if res.Student.ProfilePicture != nil {
		s.ProfilePicture = &profilePicture{
			PresignedURL: res.Student.ProfilePicture.PresignedURL,
		}
	}
	return map[string]any{
		"schoolId":              strconv.FormatInt(res.SchoolID, 10),
		"assignmentSubmissions": submissions,
		"student":               s,
	}
}

func (h *Handler) ImportStudentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !schools.Can(schools.PermissionStudentCreate, userID, schoolID, ctx, h.DB) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	type studentPayload struct {
		Name     string  `json:"name"`
		LastName string  `json:"lastName"`
		DoB      string  `json:"dateOfBirth"`
		Email    string  `json:"email"`
		Phone    *string `json:"phone"`
		Notes    *string `json:"notes"`
	}

	type payload struct {
		Students       []studentPayload `json:"students"`
		EnableAccounts bool             `json:"enableAccounts"`
	}
	var p payload

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(p.Students) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if len(p.Students) > 5_000 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staffhelpers.ErrorPayloadTooLong,
		})
		return
	}

	type student struct {
		ID                  int64
		Name                string
		LastName            string
		DoB                 time.Time
		Email               string
		Phone               *string
		Notes               *string
		AccountEnabled      bool
		ActivationToken     *string
		ActivationTokenHash *[]byte
	}
	var studentList []student

	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	for i, s := range p.Students {
		// Parse & validate
		s.Name = strings.TrimSpace(s.Name)
		s.LastName = strings.TrimSpace(s.LastName)
		s.Email = strings.TrimSpace(strings.ToLower(s.Email))
		s.DoB = strings.TrimSpace(s.DoB)
		var phone string
		var notes string
		if s.Phone != nil {
			phone = *s.Phone
		}
		if s.Notes != nil {
			notes = *s.Notes
		}
		notes = strings.TrimSpace(notes)
		phone = strings.TrimSpace(phone)
		dob, err := time.Parse("2006-01-02", s.DoB)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if s.Name == "" || len(s.Name) < 1 || len(s.Name) > 32 ||
			s.LastName == "" || len(s.LastName) < 1 || len(s.LastName) > 32 ||
			s.Email == "" || len(s.Email) < 5 || len(s.Email) > 254 ||
			phone != "" && len(phone) < 3 || len(phone) > 32 ||
			notes != "" && len(notes) > 2048 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"invalidRow": i,
			})
			return
		}

		// Generate new ID
		newID, err := h.Sf.NextID()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		sRow := student{
			ID:             newID,
			Name:           s.Name,
			LastName:       s.LastName,
			DoB:            dob,
			Email:          s.Email,
			Phone:          s.Phone,
			Notes:          s.Notes,
			AccountEnabled: p.EnableAccounts,
		}

		if p.EnableAccounts {
			// Generate activation token
			activationToken, err := gonanoid.New(128)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			activationTokenHash := helpers.MakeHash256(activationToken)

			sRow.ActivationToken = &activationToken
			sRow.ActivationTokenHash = &activationTokenHash
		}

		studentList = append(studentList, sRow)
	}

	// The number of arguments one insert query would take
	argLen := 10

	activationArgs := make([]any, 0, len(studentList)*2) // Query uses 2 parameters (portal_user_id, token_hash)
	activationPlaceholders := make([]string, 0, len(studentList))

	args := make([]any, 0, len(studentList)*argLen)
	placeholders := make([]string, 0, len(studentList))

	confArgs := make([]any, 0, len(studentList))
	confPlaceholders := make([]string, 0, len(studentList))

	var emails []string
	mailQ := make([]helpers.MailQueue, 0, len(studentList))

	for i, s := range studentList {
		n := i*argLen + 1
		an := i*2 + 1
		cn := i + 1

		placeholders = append(
			placeholders,
			fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				n, n+1, n+2, n+3, n+4, n+5, n+6, n+7, n+8, n+9),
		)

		confArgs = append(confArgs, s.Email)
		confPlaceholders = append(
			confPlaceholders,
			fmt.Sprintf("($%d)", cn),
		)

		args = append(args,
			s.ID,
			schoolID,
			s.Name,
			s.LastName,
			s.DoB,
			s.Email,
			s.Phone,
			s.Notes,
			s.AccountEnabled,
			helpers.AccTypeStudent,
		)

		// Build placeholders & args for the account activation token insert query
		activationPlaceholders = append(
			activationPlaceholders,
			fmt.Sprintf("($%d, $%d)",
				an, an+1),
		)
		activationArgs = append(activationArgs, s.ID, s.ActivationTokenHash)

		emails = append(emails, s.Email)

		if s.AccountEnabled && s.ActivationToken != nil {
			appName := os.Getenv("APP_NAME")
			frontendURL := os.Getenv("FRONTEND_URL")
			verificationURL := fmt.Sprintf("%s/app/portal/auth/activate/%s", frontendURL, *s.ActivationToken)
			msgBody := fmt.Sprintf(`
Hello %s,

Your %s Portal account has been created. Please activate it here:
%s
`,
				s.Name,
				appName,
				verificationURL,
			)
			appName = strings.TrimSpace(appName)

			mailQ = append(mailQ, helpers.MailQueue{
				To:          s.Email,
				Subject:     "Your EduLink Portal Account",
				ContentType: mail.TypeTextPlain,
				Importance:  mail.ImportanceNormal,
				Body:        msgBody,
				SendAt:      nil,
				Purpose:     helpers.Ptr(helpers.MailPurposeStudentAccountCreation),
				MaxRetries:  nil,
			})
		}
	}

	if p.EnableAccounts {
		if err := helpers.AppendMailQueue(mailQ, h.Sf, tx, ctx); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Check for email conflicts
	conflictQuery := `
SELECT email FROM portal_users WHERE email IN ($0)
`
	conflictQuery = strings.ReplaceAll(conflictQuery, "$0", strings.Join(confPlaceholders, ","))

	var conflictingEmails []string
	rows, err := tx.QueryContext(ctx, conflictQuery, confArgs...)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var e string

		if err := rows.Scan(&e); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		conflictingEmails = append(conflictingEmails, e)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(conflictingEmails) > 0 {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"conflictingEmails": conflictingEmails,
		})
		return
	}

	// Insert students

	// Workaround to suppress false SQL syntax errors
	query := `
INSERT INTO portal_users (id, school_id, name, last_name, date_of_birth, email, phone, notes, account_enabled, account_type)
VALUES ($0,$0,$0,$0,$0,$0,$0,$0,$0,$0)`
	query = strings.ReplaceAll(query, "($0,$0,$0,$0,$0,$0,$0,$0,$0,$0)", strings.Join(placeholders, ","))

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Insert verification tokens
	if p.EnableAccounts {
		query = `
INSERT INTO portal_account_activation_tokens (portal_user_id, token_hash)
VALUES ($0,$0)
`
		query = strings.ReplaceAll(query, "($0,$0)", strings.Join(activationPlaceholders, ","))

		if _, err := tx.ExecContext(ctx, query, activationArgs...); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) UploadStudentPfpHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	portalUserID, err := strconv.ParseInt(chi.URLParam(r, "studentID"), 10, 64)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	type payload struct {
		Name                string `json:"name"`
		DeclaredSize        int64  `json:"declaredSize"`
		DeclaredContentType string `json:"declaredContentType"`
	}
	var p payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res, err := h.App.Staff.Students.UploadStudentProfilePicture(ctx, &staffStudents.UploadProfilePictureInput{
		UserID:       userID,
		PortalUserID: portalUserID,
		File: staffStudents.UploadProfilePictureFile{
			FileName:            p.Name,
			DeclaredSize:        p.DeclaredSize,
			DeclaredContentType: p.DeclaredContentType,
		},
	})
	if err != nil {
		switch {
		case errors.Is(err, staffStudents.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)

		case errors.Is(err, staffStudents.ErrInvalidData):
			w.WriteHeader(http.StatusUnprocessableEntity)

		case errors.Is(err, staffStudents.ErrProfilePictureExists):
			w.WriteHeader(http.StatusConflict)

		default:
			w.WriteHeader(http.StatusInternalServerError)

		}

		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":              strconv.FormatInt(res.ID, 10),
		"completionToken": res.CompletionToken,
		"url":             res.URL,
	})
}

func (h *Handler) RemoveStudentPfpHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	portalUserID, err := strconv.ParseInt(chi.URLParam(r, "studentID"), 10, 64)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.App.Staff.Students.ClearStudentProfilePicture(ctx, &staffStudents.ClearStudentProfilePictureInput{
		UserID:       userID,
		PortalUserID: portalUserID,
	}); err != nil {
		switch {
		case errors.Is(err, staffStudents.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
