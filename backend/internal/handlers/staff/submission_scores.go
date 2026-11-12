package staff

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"app/internal/application/staff/submissionscores"
	"app/internal/auth"

	"github.com/go-chi/chi/v5"
)

type GradeSubmissionPayload struct {
	Score int    `json:"score"`
	Notes string `json:"notes"`
}

func (h *Handler) GradeSubmissionHandler(w http.ResponseWriter, r *http.Request) {
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

	var p GradeSubmissionPayload

	dec := json.NewDecoder(r.Body)
	decodeErr := dec.Decode(&p)
	err = h.App.Staff.SubmissionScores.GradeSubmission(ctx, &submissionscores.GradeSubmissionInput{InvalidPayload: decodeErr != nil, UserID: userID, SubmissionID: subID, Data: submissionscores.GradeSubmissionData{Score: p.Score, Notes: p.Notes}})
	if err != nil {
		switch {
		case errors.Is(err, submissionscores.ErrBadRequest):
			log.Println(decodeErr)
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, submissionscores.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, submissionscores.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, submissionscores.ErrConflict):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) ClearSubmissionScoreHandler(w http.ResponseWriter, r *http.Request) {
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

	err = h.App.Staff.SubmissionScores.ClearSubmissionScore(ctx, &submissionscores.ClearSubmissionScoreInput{UserID: userID, SubmissionID: subID})
	if err != nil {
		switch {
		case errors.Is(err, submissionscores.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
