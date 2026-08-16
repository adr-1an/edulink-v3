package staff

import (
	"app/internal/handlers/portal/students"
	helpers2 "app/internal/helpers"
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
	"github.com/sony/sonyflake/v2"
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

	if p.Name == "" || len(p.Name) < 3 || len(p.Name) > 32 ||
		p.LastName == "" || len(p.LastName) < 3 || len(p.LastName) > 32 ||
		p.Email == "" || len(p.Email) < 5 || len(p.Email) > 254 ||
		p.Phone != "" && len(p.Phone) < 3 || len(p.Phone) > 32 ||
		p.Notes != "" && len(p.Notes) > 2048 {
		return StudentPayload{}, errors.New("validation failed")
	}

	return p, nil
}

func checkSchoolAndEmailConflict(tx *sql.Tx, ctx context.Context, schoolID int64, email string) (bool, error) {
	var conflict bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM portal_users WHERE school_id = $1 AND email = $2 AND account_type = $3
		)
	`, schoolID, email, helpers2.AccTypeStudent).Scan(&conflict); err != nil {
		log.Println(err)
		return false, err
	}

	return conflict, nil
}

func studentToSchoolID(ctx context.Context, db *sql.DB, studentID int64) (int64, error) {
	var sID int64
	if err := db.QueryRowContext(ctx, `
		SELECT school_id FROM portal_users WHERE id = $1 AND account_type = $2
	`, studentID, helpers2.AccTypeStudent).Scan(&sID); err != nil {
		return 0, err
	}

	return sID, nil
}

func CreateStudentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
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

	if !schools.Can(schools.PermissionStudentCreate, userID, schoolID, ctx, db) {
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

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	conflict, err := checkSchoolAndEmailConflict(tx, ctx, schoolID, p.Email)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if conflict {
		w.WriteHeader(http.StatusConflict)
		return
	}

	id, err := sf.NextID()
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
			password_hash,
			account_type
			)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
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
		passHash,
		helpers2.AccTypeStudent,
	)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	details := fmt.Sprintf("{user} created a new student: %s <%s>", p.Name, p.Email)
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionStudentCreate, schools.TypeCreate, "Student created", "{user} created a student.", tx, ctx, sf, details); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func UpdateStudentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	studentID, err := strconv.ParseInt(chi.URLParam(r, "studentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	schoolID, err := studentToSchoolID(ctx, db, studentID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !schools.Can(schools.PermissionStudentUpdate, userID, schoolID, ctx, db) {
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

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

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
		lastname string
		email    string
	)
	if err := db.QueryRowContext(ctx, `
		SELECT name, last_name, email
		FROM portal_users
		WHERE id = $1
		AND school_id = $2
		AND account_type = $3
	`, studentID, schoolID, helpers2.AccTypeStudent).Scan(&name, &lastname, &email); err != nil {
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
		    password_hash = COALESCE($8, password_hash)
		WHERE id = $9
		AND school_id = $10
		AND account_type = $11
	`,
		p.Name,
		p.LastName,
		p.DoB,
		p.Email,
		p.Phone,
		p.Notes,
		p.AccountEnabled,
		passHash,
		studentID,
		schoolID,
		helpers2.AccTypeStudent,
	)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if passHash != nil {
		// Password has been changed, delete all sessions
		if _, err := db.ExecContext(ctx, `
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
		name, lastname, email, p.Name, p.LastName, p.Email)
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionStudentEdit, schools.TypeEdit, "Student updated", "{user} updated a student.", tx, ctx, sf, details); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteStudentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	studentID, err := strconv.ParseInt(chi.URLParam(r, "studentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	schoolID, err := studentToSchoolID(ctx, db, studentID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !schools.Can(schools.PermissionStudentDelete, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
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
	`, studentID, schoolID, helpers2.AccTypeStudent).Scan(&name, &email); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	details := fmt.Sprintf("{user} deleted the student %s <%s>.", name, email)
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionStudentDelete, schools.TypeDelete, "Student deleted", "{user} deleted a student.", tx, ctx, sf, details); err != nil {
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

func ListStudentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	if !schools.Can(schools.PermissionStudentList, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	type student struct {
		ID             string    `json:"id"`
		Name           string    `json:"name"`
		LastName       string    `json:"last_name"`
		DoB            *string   `json:"dateOfBirth"`
		Email          *string   `json:"email"`
		Phone          *string   `json:"phone"`
		Notes          *string   `json:"notes"`
		AccountEnabled bool      `json:"accountEnabled"`
		CreatedAt      time.Time `json:"createdAt"`
	}
	var studentList []student

	rows, err := tx.QueryContext(ctx, `
		SELECT
		    id, name, last_name, date_of_birth, email, phone, notes, account_enabled, created_at
		FROM portal_users
		WHERE school_id = $1
		AND account_type = $2
	`, schoolID, helpers2.AccTypeStudent)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var s student

		if err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.LastName,
			&s.DoB,
			&s.Email,
			&s.Phone,
			&s.Notes,
			&s.AccountEnabled,
			&s.CreatedAt,
		); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		studentList = append(studentList, s)
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
		"access":   access,
		"students": studentList,
	})
}

func ViewStudentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	studentID, err := strconv.ParseInt(chi.URLParam(r, "studentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	schoolID, err := studentToSchoolID(ctx, db, studentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if !schools.Can(schools.PermissionStudentView, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

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

	type student struct {
		ID             string    `json:"id"`
		Name           string    `json:"name"`
		Lastname       string    `json:"lastName"`
		DoB            string    `json:"dateOfBirth"`
		Email          string    `json:"email"`
		Phone          *string   `json:"phone"`
		Notes          *string   `json:"notes"`
		AccountEnabled bool      `json:"accountEnabled"`
		CreatedAt      time.Time `json:"createdAt"`
	}
	var s student

	// Get the student data & submissions
	rows, err := db.QueryContext(ctx, `
		SELECT
			s.id, s.created_at, s.notes,
			ss.score_percentage,
			a.id, a.title, a.description,
			c.id, c.name, c.color,
			g.id, g.name, g.level,
		
			st.id, st.name, st.last_name, st.date_of_birth, st.email, st.phone,
			st.notes, st.account_enabled, st.created_at
		FROM portal_users st
		LEFT JOIN assignment_submissions s
			ON s.submitted_by = st.id
			AND s.status = $1
		LEFT JOIN submission_scores ss ON ss.submission_id = s.id
		LEFT JOIN assignments a ON a.id = s.assignment_id
		LEFT JOIN courses c ON c.id = a.course_id
		LEFT JOIN grades g ON g.id = c.grade_id
		WHERE st.id = $2
	`, students.SubmissionStatusSubmitted, studentID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	var submissions []submission

	for rows.Next() {
		var sub *submission
		var subScore *int

		var (
			subID                         *string
			submittedAt                   *time.Time
			subNotes                      *string
			subAssignmentID               *string
			subAssignmentTitle            *string
			subAssignmentDescription      *string
			subAssignmentCourseID         *string
			subAssignmentCourseName       *string
			subAssignmentCourseColor      *string
			subAssignmentCourseGradeID    *string
			subAssignmentCourseGradeName  *string
			subAssignmentCourseGradeLevel *int
		)

		if err := rows.Scan(
			&subID,
			&submittedAt,
			&subNotes,

			&subScore,

			&subAssignmentID,
			&subAssignmentTitle,
			&subAssignmentDescription,

			&subAssignmentCourseID,
			&subAssignmentCourseName,
			&subAssignmentCourseColor,

			&subAssignmentCourseGradeID,
			&subAssignmentCourseGradeName,
			&subAssignmentCourseGradeLevel,

			&s.ID,
			&s.Name,
			&s.Lastname,
			&s.DoB,
			&s.Email,
			&s.Phone,
			&s.Notes,
			&s.AccountEnabled,
			&s.CreatedAt,
		); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if subID != nil && subAssignmentID != nil && subAssignmentCourseID != nil && subAssignmentCourseGradeID != nil {
			sub = &submission{
				ID: *subID,
				Assignment: assignment{
					ID: *subAssignmentID,
					Course: course{
						ID: *subAssignmentCourseID,
						Grade: grade{
							ID:    *subAssignmentCourseGradeID,
							Name:  *subAssignmentCourseGradeName,
							Level: *subAssignmentCourseGradeLevel,
						},
						Name:  *subAssignmentCourseName,
						Color: *subAssignmentCourseColor,
					},
					Title:       *subAssignmentTitle,
					Description: subAssignmentDescription,
				},
				Score:       nil,
				SubmittedAt: *submittedAt,
				Notes:       subNotes,
			}

			if subScore != nil {
				sub.Score = &assignmentScore{
					ScorePercentage: *subScore,
				}
			}
			submissions = append(submissions, *sub)
		}
	}

	if s.Notes != nil && *s.Notes == "" {
		s.Notes = nil
	}

	access, err := schools.GetAllUserPermissions(ctx, db, userID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":                access,
		"assignmentSubmissions": submissions,
		"student":               s,
	})
}

func ImportStudentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !schools.Can(schools.PermissionStudentCreate, userID, schoolID, ctx, db) {
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
			"code": staff_helpers.ErrorPayloadTooLong,
		})
		return
	}

	type studentCredentials struct {
		RawPassword  string
		PasswordHash string
	}

	type student struct {
		ID             int64
		Name           string
		LastName       string
		DoB            time.Time
		Email          string
		Phone          *string
		Notes          *string
		AccountEnabled bool
		Credentials    *studentCredentials
	}
	var studentList []student

	tx, err := db.BeginTx(ctx, nil)
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

		if s.Name == "" || len(s.Name) < 3 || len(s.Name) > 32 ||
			s.LastName == "" || len(s.LastName) < 3 || len(s.LastName) > 32 ||
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
		newID, err := sf.NextID()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var creds *studentCredentials
		if p.EnableAccounts {
			// Generate password
			pass, err := gonanoid.New(16)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			passHash, err := argon2id.CreateHash(pass, argon2id.DefaultParams)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			creds = &studentCredentials{
				RawPassword:  pass,
				PasswordHash: passHash,
			}
		}

		studentList = append(studentList, student{
			ID:             newID,
			Name:           s.Name,
			LastName:       s.LastName,
			DoB:            dob,
			Email:          s.Email,
			Phone:          s.Phone,
			Notes:          s.Notes,
			AccountEnabled: p.EnableAccounts,
			Credentials:    creds,
		})
	}

	args := make([]any, 0, len(studentList)*11)
	placeholders := make([]string, 0, len(studentList))

	confArgs := make([]any, 0, len(studentList))
	confPlaceholders := make([]string, 0, len(studentList))

	var emails []string
	mailQ := make([]helpers2.MailQueue, 0, len(emails))

	for i, s := range studentList {
		n := i*11 + 1
		cn := i + 1

		placeholders = append(
			placeholders,
			fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				n, n+1, n+2, n+3, n+4, n+5, n+6, n+7, n+8, n+9, n+10),
		)

		confArgs = append(confArgs, s.Email)
		confPlaceholders = append(
			confPlaceholders,
			fmt.Sprintf("($%d)", cn),
		)

		var passHash *string
		if s.Credentials != nil {
			passHash = &s.Credentials.PasswordHash
		}

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
			passHash,
			helpers2.AccTypeStudent,
		)

		emails = append(emails, s.Email)

		if s.AccountEnabled && s.Credentials != nil {
			appName := os.Getenv("APP_NAME")
			msgBody := fmt.Sprintf(`
Hello %s,

Your %s Portal account has been created.
Here are your credentials:

Email: %s
Password: %s

Please change your password as soon as you log in.
`,
				s.Name,
				appName,
				s.Email,
				s.Credentials.RawPassword,
			)
			appName = strings.TrimSpace(appName)

			mailQ = append(mailQ, helpers2.MailQueue{
				To:          s.Email,
				Subject:     "Your EduLink Portal Account",
				ContentType: mail.TypeTextPlain,
				Importance:  mail.ImportanceNormal,
				Body:        msgBody,
				SendAt:      nil,
				Purpose:     helpers2.Ptr(helpers2.MailPurposeStudentAccountCreation),
				MaxRetries:  nil,
			})

			if err := helpers2.AppendMailQueue(mailQ, sf, tx, ctx); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
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

	// Workaround to suppress false SQL syntax errors
	query := `
INSERT INTO portal_users (id, school_id, name, last_name, date_of_birth, email, phone, notes, account_enabled, password_hash, account_type)
VALUES ($0,$0,$0,$0,$0,$0,$0,$0,$0,$0,$0)`
	query = strings.ReplaceAll(query, "($0,$0,$0,$0,$0,$0,$0,$0,$0,$0,$0)", strings.Join(placeholders, ","))

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
