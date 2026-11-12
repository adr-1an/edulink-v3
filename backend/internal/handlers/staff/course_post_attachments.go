package staff

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"app/internal/application/staff/coursepostattachments"
	"app/internal/helpers/staff"

	"github.com/go-chi/chi/v5"
)

type InitPostAttachmentUploadPayload struct {
	Name                string `json:"name"`
	DeclaredSize        int64  `json:"declaredSize"`
	DeclaredContentType string `json:"declaredContentType"`
}

func (h *Handler) InitPostAttachmentUploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	postID, err := strconv.ParseInt(chi.URLParam(r, "postID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var p InitPostAttachmentUploadPayload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)
	res, err := h.App.Staff.CoursePostAttachments.InitPostAttachmentUpload(ctx, &coursepostattachments.InitPostAttachmentUploadInput{InvalidPayload: decodeErr != nil, UserID: userID, PostID: postID, Data: coursepostattachments.InitPostAttachmentUploadData{Name: p.Name, DeclaredSize: p.DeclaredSize, DeclaredContentType: p.DeclaredContentType}})
	if err != nil {
		switch {
		case errors.Is(err, coursepostattachments.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, coursepostattachments.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, coursepostattachments.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
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

func (h *Handler) DeletePostAttachmentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	attachmentID, err := strconv.ParseInt(chi.URLParam(r, "attachmentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.App.Staff.CoursePostAttachments.DeletePostAttachment(ctx, &coursepostattachments.DeletePostAttachmentInput{UserID: userID, AttachmentID: attachmentID})
	if err != nil {
		switch {
		case errors.Is(err, coursepostattachments.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
