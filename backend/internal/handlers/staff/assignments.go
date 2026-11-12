package staff

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"app/internal/application/staff/assignments"
	"app/internal/auth"

	"github.com/go-chi/chi/v5"
)

type CreateAssignmentPayload struct {
	RefPostID          *string    `json:"referencedPostId"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	DueDate            *time.Time `json:"dueDate"`
	SubmissionsEnabled bool       `json:"submissionsEnabled"`
	SubmissionsCloseAt *time.Time `json:"submissionsCloseAt"`
	Archived           bool       `json:"archived"`
}

func (h *Handler) CreateAssignmentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var p CreateAssignmentPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)
	err = h.App.Staff.Assignments.CreateAssignment(ctx, &assignments.CreateAssignmentInput{InvalidPayload: decodeErr != nil, UserID: userID, CourseID: courseID, Data: assignments.Assignment{RefPostID: p.RefPostID, Title: p.Title, Description: p.Description, DueDate: p.DueDate, SubmissionsEnabled: p.SubmissionsEnabled, SubmissionsCloseAt: p.SubmissionsCloseAt, Archived: p.Archived}})
	if err != nil {
		switch {
		case errors.Is(err, assignments.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, assignments.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, assignments.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
}

type ListAssignmentsPost struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func mapListAssignmentsPost(v assignments.ListAssignmentsPost) ListAssignmentsPost {
	out := ListAssignmentsPost{
		ID:    v.ID,
		Title: v.Title,
	}
	return out
}

type ListAssignmentsAssignment struct {
	RefPost            *ListAssignmentsPost `json:"referencedPost"`
	ID                 string               `json:"id"`
	Title              string               `json:"title"`
	Description        string               `json:"description"`
	DueDate            *time.Time           `json:"dueDate"`
	SubmissionsEnabled bool                 `json:"submissionsEnabled"`
	SubmissionsCloseAt *time.Time           `json:"submissionsCloseAt"`
	CreatedAt          time.Time            `json:"createdAt"`
	ArchivedAt         *time.Time           `json:"archivedAt"`
}

func mapListAssignmentsAssignment(v assignments.ListAssignmentsAssignment) ListAssignmentsAssignment {
	out := ListAssignmentsAssignment{
		ID:                 v.ID,
		Title:              v.Title,
		Description:        v.Description,
		DueDate:            v.DueDate,
		SubmissionsEnabled: v.SubmissionsEnabled,
		SubmissionsCloseAt: v.SubmissionsCloseAt,
		CreatedAt:          v.CreatedAt,
		ArchivedAt:         v.ArchivedAt,
	}
	if v.RefPost != nil {
		mapped := mapListAssignmentsPost(*v.RefPost)
		out.RefPost = &mapped
	}
	return out
}

func (h *Handler) ListAssignmentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Assignments.ListAssignments(ctx, &assignments.ListAssignmentsInput{UserID: userID, CourseID: courseID})
	if err != nil {
		switch {
		case errors.Is(err, assignments.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	var mappedAssignments []ListAssignmentsAssignment
	if res.Assignments != nil {
		mappedAssignments = make([]ListAssignmentsAssignment, 0, len(res.Assignments))
		for _, item := range res.Assignments {
			mappedAssignments = append(mappedAssignments, mapListAssignmentsAssignment(item))
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":      res.Access,
		"assignments": mappedAssignments,
	})
}

type UpdateAssignmentPayload struct {
	RefPostID          *string    `json:"referencedPostId"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	DueDate            *time.Time `json:"dueDate"`
	SubmissionsEnabled bool       `json:"submissionsEnabled"`
	SubmissionsCloseAt *time.Time `json:"submissionsCloseAt"`
	Archived           bool       `json:"archived"`
}

func (h *Handler) UpdateAssignmentHandler(w http.ResponseWriter, r *http.Request) {
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

	var p UpdateAssignmentPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)
	err = h.App.Staff.Assignments.UpdateAssignment(ctx, &assignments.UpdateAssignmentInput{InvalidPayload: decodeErr != nil, UserID: userID, AssignmentID: assignmentID, Data: assignments.Assignment{RefPostID: p.RefPostID, Title: p.Title, Description: p.Description, DueDate: p.DueDate, SubmissionsEnabled: p.SubmissionsEnabled, SubmissionsCloseAt: p.SubmissionsCloseAt, Archived: p.Archived}})
	if err != nil {
		switch {
		case errors.Is(err, assignments.ErrBadRequest):
			log.Println(decodeErr)
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, assignments.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, assignments.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteAssignmentHandler(w http.ResponseWriter, r *http.Request) {
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

	err = h.App.Staff.Assignments.DeleteAssignment(ctx, &assignments.DeleteAssignmentInput{UserID: userID, AssignmentID: assignmentID})
	if err != nil {
		switch {
		case errors.Is(err, assignments.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
