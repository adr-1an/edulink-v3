package students

import (
	svc "app/internal/application/portal/students"
	"app/internal/auth"
	"app/internal/portalusers"

	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) DeleteSubmissionAttachment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUID(w, r, h.DB, ctx, portalusers.AccTypeStudent)
	if err != nil {
		return
	}

	submissionID, err := strconv.ParseInt(chi.URLParam(r, "submissionID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	attachmentID, err := strconv.ParseInt(chi.URLParam(r, "attachmentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.App.Portal.Students.DeleteSubmissionAttachment(ctx, &svc.DeleteSubmissionAttachmentInput{UserID: userID, SubmissionID: submissionID, AttachmentID: attachmentID})
	if err != nil {
		writeStudentsError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) CreateAssignmentSubmission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUID(w, r, h.DB, ctx, portalusers.AccTypeStudent)
	if err != nil {
		return
	}

	assignmentID, err := strconv.ParseInt(chi.URLParam(r, "assignmentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	type payload struct {
		Notes string `json:"notes"`
	}
	var p payload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)

	res, err := h.App.Portal.Students.CreateAssignmentSubmission(ctx, &svc.CreateAssignmentSubmissionInput{UserID: userID, AssignmentID: assignmentID, InvalidPayload: decodeErr != nil, Notes: p.Notes})
	if err != nil {
		writeStudentsError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"submissionId": res.SubmissionID, "status": res.Status})
}
func (h *Handler) InitSubmissionAttachmentUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUID(w, r, h.DB, ctx, portalusers.AccTypeStudent)
	if err != nil {
		return
	}

	submissionID, err := strconv.ParseInt(chi.URLParam(r, "submissionID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	type payload struct {
		Name                string `json:"fileName"`
		DeclaredSize        int64  `json:"declaredSize"`
		DeclaredContentType string `json:"declaredContentType"`
	}
	var p payload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)

	res, err := h.App.Portal.Students.InitSubmissionAttachmentUpload(ctx, &svc.InitSubmissionAttachmentUploadInput{UserID: userID, SubmissionID: submissionID, InvalidPayload: decodeErr != nil, Name: p.Name, DeclaredSize: p.DeclaredSize, DeclaredContentType: p.DeclaredContentType})
	if err != nil {
		writeStudentsError(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"id": res.ID, "attachmentId": res.AttachmentID, "completionToken": res.CompletionToken, "url": res.URL})
}
func (h *Handler) CompleteSubmissionCreation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUID(w, r, h.DB, ctx, portalusers.AccTypeStudent)
	if err != nil {
		return
	}

	submissionID, err := strconv.ParseInt(chi.URLParam(r, "submissionID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.App.Portal.Students.CompleteSubmissionCreation(ctx, &svc.CompleteSubmissionCreationInput{UserID: userID, SubmissionID: submissionID})
	if err != nil {
		writeStudentsError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
