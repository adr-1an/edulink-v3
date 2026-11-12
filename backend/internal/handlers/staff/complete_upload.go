package staff

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"app/internal/application/staff/uploads"
	"app/internal/auth"

	"github.com/go-chi/chi/v5"
)

type CompleteUploadPayload struct {
	CompletionToken string `json:"completionToken"`
}

func (h *Handler) CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	objectID, err := strconv.ParseInt(chi.URLParam(r, "objectID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var p CompleteUploadPayload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.App.Staff.Uploads.CompleteUpload(ctx, &uploads.CompleteUploadInput{UserID: userID, ObjectID: objectID, Data: uploads.CompleteUploadData{CompletionToken: p.CompletionToken}})
	if err != nil {
		switch {
		case errors.Is(err, uploads.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, uploads.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
