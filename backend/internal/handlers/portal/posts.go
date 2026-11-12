package portal

import (
	svc "app/internal/application/portal/posts"
	"app/internal/helpers"
	portal "app/internal/helpers/portal"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ViewPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := portal.TokenToUID(w, r, h.DB, ctx, helpers.AccTypeEither)
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
