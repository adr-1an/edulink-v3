package staff

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	staffStudents "app/internal/application/staff/students"
	staffhelpers "app/internal/helpers/staff"

	"github.com/go-chi/chi/v5"
)

type StudentPayload = staffStudents.StudentPayload

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

	var p StudentPayload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	invalidPayload := dec.Decode(&p) != nil
	err = h.App.Staff.Students.CreateStudent(ctx, &staffStudents.CreateStudentInput{InvalidPayload: invalidPayload, UserID: userID, SchoolID: schoolID, Student: p})
	if err != nil {
		switch {
		case errors.Is(err, staffStudents.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, staffStudents.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, staffStudents.ErrInvalidData):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, staffStudents.ErrConflict):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
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

	var p StudentPayload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	invalidPayload := dec.Decode(&p) != nil
	err = h.App.Staff.Students.UpdateStudent(ctx, &staffStudents.UpdateStudentInput{InvalidPayload: invalidPayload, UserID: userID, StudentID: studentID, Student: p})
	if err != nil {
		switch {
		case errors.Is(err, staffStudents.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, staffStudents.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, staffStudents.ErrInvalidData):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, staffStudents.ErrConflict):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
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
	err = h.App.Staff.Students.DeleteStudent(ctx, &staffStudents.DeleteStudentInput{UserID: userID, StudentID: studentID})
	if err != nil {
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

	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":   res.Access,
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

	payload := studentViewPayload(res)
	payload["access"] = res.Access
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

	var p staffStudents.ImportStudentsData
	invalidPayload := json.NewDecoder(r.Body).Decode(&p) != nil
	res, err := h.App.Staff.Students.ImportStudents(ctx, &staffStudents.ImportStudentsInput{InvalidPayload: invalidPayload, UserID: userID, SchoolID: schoolID, Data: p})
	if err != nil {
		var invalidRow *staffStudents.InvalidRowError
		var emailConflict *staffStudents.EmailConflictError
		switch {
		case errors.Is(err, staffStudents.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, staffStudents.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, staffStudents.ErrInvalidData):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, staffStudents.ErrConflict):
			w.WriteHeader(http.StatusConflict)
		case errors.Is(err, staffStudents.ErrPayloadTooLong):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": staffhelpers.ErrorPayloadTooLong})
		case errors.As(err, &invalidRow):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{"invalidRow": invalidRow.Row})
		case errors.As(err, &emailConflict):
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{"conflictingEmails": emailConflict.Emails})
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if !res.Created {
		w.WriteHeader(http.StatusNoContent)
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
