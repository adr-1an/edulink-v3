package portal

import (
	svc "app/internal/application/portal/posts"
	"app/internal/auth"
	"app/internal/portalusers"

	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ViewPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUID(w, r, h.DB, ctx, portalusers.AccTypeEither)
	if err != nil {
		return
	}

	postID, err := strconv.ParseInt(chi.URLParam(r, "postID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res, err := h.App.Portal.Posts.ViewPost(ctx, &svc.ViewPostInput{UserID: userID, PostID: postID})
	if err != nil {
		writePostsError(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"post": mapViewPostPost(res.Post)})
}
