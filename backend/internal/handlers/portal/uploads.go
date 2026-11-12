package portal

import (
	svc "app/internal/application/portal/uploads"
	"app/internal/auth"
	"app/internal/portalusers"

	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) CompleteUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUID(w, r, h.DB, ctx, portalusers.AccTypeEither)
	if err != nil {
		return
	}

	objectID, err := strconv.ParseInt(chi.URLParam(r, "objectID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	type payload struct {
		CompletionToken string `json:"completionToken"`
	}
	var p payload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)

	err = h.App.Portal.Uploads.CompleteUpload(ctx, &svc.CompleteUploadInput{UserID: userID, ObjectID: objectID, InvalidPayload: decodeErr != nil, CompletionToken: p.CompletionToken})
	if err != nil {
		writeUploadsError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
