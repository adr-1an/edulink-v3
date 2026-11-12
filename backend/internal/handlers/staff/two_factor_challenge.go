package staff

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"app/internal/application/staff/auth"
	"app/internal/responsecodes"
)

type challengeCompletionPayload struct {
	Code           string `json:"code"`
	ChallengeToken string `json:"challengeToken"`
}

func (h *Handler) CompleteTwoFactorChallenge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var p challengeCompletionPayload

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Auth.CompleteTwoFactorChallenge(ctx, &auth.CompleteChallengeInput{
		Code:           p.Code,
		ChallengeToken: p.ChallengeToken,
	})

	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidCode):
			w.WriteHeader(http.StatusUnprocessableEntity)

		case errors.Is(err, auth.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)

		case errors.Is(err, auth.ErrExpiredChallenge):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": responsecodes.ErrorCodeExpiredToken,
			})

		case errors.Is(err, auth.ErrInvalidTwoFactorCode):
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": responsecodes.ErrorCodeInvalidToken,
			})

		case errors.Is(err, auth.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)

		case errors.Is(err, auth.ErrUnsupportedChallengePurpose):
			w.WriteHeader(http.StatusNotImplemented)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	switch res.Type {
	case auth.CompleteChallengeLogin:
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token": res.SessionToken,
		})

	case auth.CompleteChallengeSchoolDeletion:
		w.WriteHeader(http.StatusNoContent)
	}
}
