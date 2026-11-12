package staff

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"app/internal/application/staff/submissions"
	"app/internal/auth"

	"github.com/go-chi/chi/v5"
)

type ListAssignmentSubmissionsPortalUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Name     string `json:"name"`
	Lastname string `json:"lastName"`
}

func mapListAssignmentSubmissionsPortalUser(v submissions.ListAssignmentSubmissionsPortalUser) ListAssignmentSubmissionsPortalUser {
	out := ListAssignmentSubmissionsPortalUser{
		ID:       v.ID,
		Email:    v.Email,
		Phone:    v.Phone,
		Name:     v.Name,
		Lastname: v.Lastname,
	}
	return out
}

type ListAssignmentSubmissionsSubmission struct {
	ID          string                              `json:"id"`
	Status      string                              `json:"status"`
	SubmittedBy ListAssignmentSubmissionsPortalUser `json:"submittedBy"`
	Notes       *string                             `json:"notes"`
	SubmittedAt time.Time                           `json:"submittedAt"`
	Graded      bool                                `json:"graded"`
}

func mapListAssignmentSubmissionsSubmission(v submissions.ListAssignmentSubmissionsSubmission) ListAssignmentSubmissionsSubmission {
	out := ListAssignmentSubmissionsSubmission{
		ID:          v.ID,
		Status:      v.Status,
		SubmittedBy: mapListAssignmentSubmissionsPortalUser(v.SubmittedBy),
		Notes:       v.Notes,
		SubmittedAt: v.SubmittedAt,
		Graded:      v.Graded,
	}
	return out
}

func (h *Handler) ListAssignmentSubmissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	assignmentID, err := strconv.ParseInt(chi.URLParam(r, "assignmentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Submissions.ListAssignmentSubmissions(ctx, &submissions.ListAssignmentSubmissionsInput{UserID: userID, AssignmentID: assignmentID})
	if err != nil {
		switch {
		case errors.Is(err, submissions.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	var mappedSubmissions []ListAssignmentSubmissionsSubmission
	if res.Submissions != nil {
		mappedSubmissions = make([]ListAssignmentSubmissionsSubmission, 0, len(res.Submissions))
		for _, item := range res.Submissions {
			mappedSubmissions = append(mappedSubmissions, mapListAssignmentSubmissionsSubmission(item))
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":      res.Access,
		"submissions": mappedSubmissions,
	})
}

type ViewSubmissionUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

func mapViewSubmissionUser(v submissions.ViewSubmissionUser) ViewSubmissionUser {
	out := ViewSubmissionUser{
		ID:    v.ID,
		Name:  v.Name,
		Email: v.Email,
		Phone: v.Phone,
	}
	return out
}

type ViewSubmissionGrade struct {
	Score    int                `json:"score"`
	Notes    *string            `json:"notes"`
	GradedAt time.Time          `json:"gradedAt"`
	GradedBy ViewSubmissionUser `json:"gradedBy"`
}

func mapViewSubmissionGrade(v submissions.ViewSubmissionGrade) ViewSubmissionGrade {
	out := ViewSubmissionGrade{
		Score:    v.Score,
		Notes:    v.Notes,
		GradedAt: v.GradedAt,
		GradedBy: mapViewSubmissionUser(v.GradedBy),
	}
	return out
}

type ViewSubmissionPortalUser struct {
	ID       string  `json:"id"`
	Email    string  `json:"email"`
	Phone    *string `json:"phone"`
	Name     string  `json:"name"`
	Lastname string  `json:"lastName"`
}

func mapViewSubmissionPortalUser(v submissions.ViewSubmissionPortalUser) ViewSubmissionPortalUser {
	out := ViewSubmissionPortalUser{
		ID:       v.ID,
		Email:    v.Email,
		Phone:    v.Phone,
		Name:     v.Name,
		Lastname: v.Lastname,
	}
	return out
}

type ViewSubmissionSubmissionAttachment struct {
	ID           string `json:"id"`
	PresignedURL string `json:"presignedUrl"`
	Filename     string `json:"originalFilename"`
	ContentType  string `json:"declaredContentType"`
}

func mapViewSubmissionSubmissionAttachment(v submissions.ViewSubmissionSubmissionAttachment) ViewSubmissionSubmissionAttachment {
	out := ViewSubmissionSubmissionAttachment{
		ID:           v.ID,
		PresignedURL: v.PresignedURL,
		Filename:     v.Filename,
		ContentType:  v.ContentType,
	}
	return out
}

type ViewSubmissionSubmission struct {
	ID          string                               `json:"id"`
	Status      string                               `json:"status"`
	SubmittedBy ViewSubmissionPortalUser             `json:"submittedBy"`
	Attachments []ViewSubmissionSubmissionAttachment `json:"attachments"`
	Grade       *ViewSubmissionGrade                 `json:"grade"`
	Notes       *string                              `json:"notes"`
	SubmittedAt time.Time                            `json:"submittedAt"`
}

func mapViewSubmissionSubmission(v submissions.ViewSubmissionSubmission) ViewSubmissionSubmission {
	out := ViewSubmissionSubmission{
		ID:          v.ID,
		Status:      v.Status,
		SubmittedBy: mapViewSubmissionPortalUser(v.SubmittedBy),
		Notes:       v.Notes,
		SubmittedAt: v.SubmittedAt,
	}
	if v.Attachments != nil {
		out.Attachments = make([]ViewSubmissionSubmissionAttachment, 0, len(v.Attachments))
		for _, item := range v.Attachments {
			out.Attachments = append(out.Attachments, mapViewSubmissionSubmissionAttachment(item))
		}
	}
	if v.Grade != nil {
		mapped := mapViewSubmissionGrade(*v.Grade)
		out.Grade = &mapped
	}
	return out
}

func (h *Handler) ViewSubmission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	subID, err := strconv.ParseInt(chi.URLParam(r, "submissionID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Submissions.ViewSubmission(ctx, &submissions.ViewSubmissionInput{UserID: userID, SubmissionID: subID})
	if err != nil {
		switch {
		case errors.Is(err, submissions.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":     res.Access,
		"submission": mapViewSubmissionSubmission(res.Submission),
	})
}

func (h *Handler) ReturnSubmission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	subID, err := strconv.ParseInt(chi.URLParam(r, "submissionID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.App.Staff.Submissions.ReturnSubmission(ctx, &submissions.ReturnSubmissionInput{UserID: userID, SubmissionID: subID})
	if err != nil {
		switch {
		case errors.Is(err, submissions.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, submissions.ErrConflict):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteReturnedSubmission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	subID, err := strconv.ParseInt(chi.URLParam(r, "submissionID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.App.Staff.Submissions.DeleteReturnedSubmission(ctx, &submissions.DeleteReturnedSubmissionInput{UserID: userID, SubmissionID: subID})
	if err != nil {
		switch {
		case errors.Is(err, submissions.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, submissions.ErrConflict):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
