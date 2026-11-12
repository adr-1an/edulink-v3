package students

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"app/internal/helpers"
	"app/internal/helpers/staff/schools"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/wneessen/go-mail"
)

type ImportStudentsInput struct {
	InvalidPayload bool
	UserID         int64
	SchoolID       int64
	Data           ImportStudentsData
}

type ImportStudent struct {
	Name     string  `json:"name"`
	LastName string  `json:"lastName"`
	DoB      string  `json:"dateOfBirth"`
	Email    string  `json:"email"`
	Phone    *string `json:"phone"`
	Notes    *string `json:"notes"`
}

type ImportStudentsData struct {
	Students       []ImportStudent `json:"students"`
	EnableAccounts bool            `json:"enableAccounts"`
}
type ImportStudentsResult struct{ Created bool }

func (svc *Service) ImportStudents(ctx context.Context, i *ImportStudentsInput) (*ImportStudentsResult, error) {
	userID := i.UserID
	schoolID := i.SchoolID
	p := i.Data

	if !schools.Can(schools.PermissionStudentCreate, userID, schoolID, ctx, svc.DB) {
		return nil, ErrForbidden
	}

	if i.InvalidPayload {
		return nil, ErrBadRequest
	}

	if len(p.Students) == 0 {
		return &ImportStudentsResult{}, nil
	}
	if len(p.Students) > 5_000 {
		return nil, ErrPayloadTooLong
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

	tx, err := svc.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
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
			return nil, ErrInvalidData
		}

		if s.Name == "" || len(s.Name) < 1 || len(s.Name) > 32 ||
			s.LastName == "" || len(s.LastName) < 1 || len(s.LastName) > 32 ||
			s.Email == "" || len(s.Email) < 5 || len(s.Email) > 254 ||
			phone != "" && len(phone) < 3 || len(phone) > 32 ||
			notes != "" && len(notes) > 2048 {
			return nil, &InvalidRowError{Row: i}
		}

		// Generate new ID
		newID, err := svc.Sf.NextID()
		if err != nil {
			log.Println(err)
			return nil, err
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
				return nil, err
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

			mailQ = append(mailQ, helpers.MailQueue{
				To:          s.Email,
				Subject:     fmt.Sprintf("Your %s account", strings.TrimSpace(appName)),
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
		if err := helpers.AppendMailQueue(mailQ, svc.Sf, tx, ctx); err != nil {
			log.Println(err)
			return nil, err
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
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var e string

		if err := rows.Scan(&e); err != nil {
			log.Println(err)
			return nil, err
		}

		conflictingEmails = append(conflictingEmails, e)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	if len(conflictingEmails) > 0 {
		return nil, &EmailConflictError{Emails: conflictingEmails}
	}

	// Insert students

	// Workaround to suppress false SQL syntax errors
	query := `
INSERT INTO portal_users (id, school_id, name, last_name, date_of_birth, email, phone, notes, account_enabled, account_type)
VALUES ($0,$0,$0,$0,$0,$0,$0,$0,$0,$0)`
	query = strings.ReplaceAll(query, "($0,$0,$0,$0,$0,$0,$0,$0,$0,$0)", strings.Join(placeholders, ","))

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		log.Println(err)
		return nil, err
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
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &ImportStudentsResult{Created: true}, nil
}
