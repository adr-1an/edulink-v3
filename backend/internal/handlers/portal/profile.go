package portal

import (
	portalProfile "app/internal/application/portal/profile"
	"app/internal/helpers"
	"app/internal/helpers/portal"
	"encoding/json"
	"net/http"
	"time"
)

func (h *Handler) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := portal.TokenToUID(w, r, h.DB, ctx, helpers.AccTypeEither)
	if err != nil {
		return
	}

	type user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type school struct {
		Owner  user   `json:"owner"`
		Name   string `json:"name"`
		Region string `json:"region"`
	}

	type profile struct {
		PfpURL      *string   `json:"pfpUrl"`
		Name        string    `json:"name"`
		LastName    string    `json:"lastName"`
		Email       string    `json:"email"`
		Phone       string    `json:"phone"`
		DateOfBirth time.Time `json:"dateOfBirth"`

		School school `json:"school"`
	}

	res, err := h.App.Portal.Profile.ViewProfile(ctx, &portalProfile.ViewProfileInput{
		UserID: userID,
	})
	if err != nil {
		switch {
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	p := &profile{
		Name:        res.Profile.Name,
		LastName:    res.Profile.LastName,
		PfpURL:      res.Profile.ProfilePictureURL,
		Email:       res.Profile.Email,
		Phone:       res.Profile.Phone,
		DateOfBirth: res.Profile.DateOfBirth,

		School: school{
			Owner: user{
				Name:  res.Profile.School.Owner.Name,
				Email: res.Profile.School.Owner.Email,
			},

			Name:   res.Profile.School.Name,
			Region: res.Profile.School.Region,
		},
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"profile": p,
	})
}
