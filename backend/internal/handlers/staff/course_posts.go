package staff

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"app/internal/application/staff/courseposts"
	"app/internal/helpers/staff"

	"github.com/go-chi/chi/v5"
)

type CreatePostPayload struct {
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	AccentColor string     `json:"accentColor"`
	ShowUntil   *time.Time `json:"showUntil"`
}

func (h *Handler) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var p CreatePostPayload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)
	res, err := h.App.Staff.CoursePosts.CreatePost(ctx, &courseposts.CreatePostInput{InvalidPayload: decodeErr != nil, UserID: userID, CourseID: courseID, Data: courseposts.Post{Title: p.Title, Body: p.Body, AccentColor: p.AccentColor, ShowUntil: p.ShowUntil}})
	if err != nil {
		switch {
		case errors.Is(err, courseposts.ErrBadRequest):
			log.Println(decodeErr)
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, courseposts.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, courseposts.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if res.AffectedRows == 0 {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id": strconv.FormatInt(res.ID, 10),
	})
}

type UpdatePostPayload struct {
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	AccentColor string     `json:"accentColor"`
	ShowUntil   *time.Time `json:"showUntil"`
}

func (h *Handler) UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
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

	var p UpdatePostPayload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)
	res, err := h.App.Staff.CoursePosts.UpdatePost(ctx, &courseposts.UpdatePostInput{InvalidPayload: decodeErr != nil, UserID: userID, PostID: postID, Data: courseposts.Post{Title: p.Title, Body: p.Body, AccentColor: p.AccentColor, ShowUntil: p.ShowUntil}})
	if err != nil {
		switch {
		case errors.Is(err, courseposts.ErrBadRequest):
			log.Println(decodeErr)
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, courseposts.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, courseposts.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if res.AffectedRows == 0 {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusNoContent)
}

type ListPostsAttachment struct {
	ID           string `json:"id"`
	PresignedURL string `json:"presignedUrl"`
	Filename     string `json:"fileName"`
	ContentType  string `json:"contentType"`
}

func mapListPostsAttachment(v courseposts.ListPostsAttachment) ListPostsAttachment {
	out := ListPostsAttachment{
		ID:           v.ID,
		PresignedURL: v.PresignedURL,
		Filename:     v.Filename,
		ContentType:  v.ContentType,
	}
	return out
}

type ListPostsPost struct {
	Attachments  []ListPostsAttachment `json:"attachments"`
	ID           string                `json:"id"`
	AuthorName   string                `json:"authorName"`
	AuthorPfpURL *string               `json:"authorProfilePictureURL"`

	Title       string     `json:"title"`
	Body        string     `json:"body"`
	ShowUntil   *time.Time `json:"showUntil"`
	AccentColor string     `json:"accentColor"`

	EditedAt  *time.Time `json:"editedAt"`
	CreatedAt time.Time  `json:"createdAt"`
}

func mapListPostsPost(v courseposts.ListPostsPost) ListPostsPost {
	out := ListPostsPost{
		ID:           v.ID,
		AuthorName:   v.AuthorName,
		AuthorPfpURL: v.AuthorPfpURL,
		Title:        v.Title,
		Body:         v.Body,
		ShowUntil:    v.ShowUntil,
		AccentColor:  v.AccentColor,
		EditedAt:     v.EditedAt,
		CreatedAt:    v.CreatedAt,
	}
	if v.Attachments != nil {
		out.Attachments = make([]ListPostsAttachment, 0, len(v.Attachments))
		for _, item := range v.Attachments {
			out.Attachments = append(out.Attachments, mapListPostsAttachment(item))
		}
	}
	return out
}

func (h *Handler) ListPostsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.CoursePosts.ListPosts(ctx, &courseposts.ListPostsInput{UserID: userID, CourseID: courseID})
	if err != nil {
		switch {
		case errors.Is(err, courseposts.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	var mappedPosts []ListPostsPost
	if res.Posts != nil {
		mappedPosts = make([]ListPostsPost, 0, len(res.Posts))
		for _, item := range res.Posts {
			mappedPosts = append(mappedPosts, mapListPostsPost(item))
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access": res.Access,
		"posts":  mappedPosts,
	})
}

func (h *Handler) DeletePostHandler(w http.ResponseWriter, r *http.Request) {
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

	res, err := h.App.Staff.CoursePosts.DeletePost(ctx, &courseposts.DeletePostInput{UserID: userID, PostID: postID})
	if err != nil {
		switch {
		case errors.Is(err, courseposts.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if res.AffectedRows == 0 {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusNoContent)
}

type ViewPostUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func mapViewPostUser(v courseposts.ViewPostUser) ViewPostUser {
	out := ViewPostUser{
		ID:    v.ID,
		Name:  v.Name,
		Email: v.Email,
	}
	return out
}

type ViewPostAttachment struct {
	ID           string `json:"id"`
	PresignedURL string `json:"presignedUrl"`
	Filename     string `json:"fileName"`
	ContentType  string `json:"contentType"`
}

func mapViewPostAttachment(v courseposts.ViewPostAttachment) ViewPostAttachment {
	out := ViewPostAttachment{
		ID:           v.ID,
		PresignedURL: v.PresignedURL,
		Filename:     v.Filename,
		ContentType:  v.ContentType,
	}
	return out
}

type ViewPostPost struct {
	Attachments []ViewPostAttachment `json:"attachments"`
	Author      ViewPostUser         `json:"author"`
	ID          string               `json:"id"`
	Title       string               `json:"title"`
	Body        string               `json:"body"`
	Color       string               `json:"accentColor"`
	ShowUntil   *time.Time           `json:"showUntil"`
	CreatedAt   time.Time            `json:"createdAt"`
	EditedAt    *time.Time           `json:"editedAt"`
}

func mapViewPostPost(v courseposts.ViewPostPost) ViewPostPost {
	out := ViewPostPost{
		Author:    mapViewPostUser(v.Author),
		ID:        v.ID,
		Title:     v.Title,
		Body:      v.Body,
		Color:     v.Color,
		ShowUntil: v.ShowUntil,
		CreatedAt: v.CreatedAt,
		EditedAt:  v.EditedAt,
	}
	if v.Attachments != nil {
		out.Attachments = make([]ViewPostAttachment, 0, len(v.Attachments))
		for _, item := range v.Attachments {
			out.Attachments = append(out.Attachments, mapViewPostAttachment(item))
		}
	}
	return out
}

func (h *Handler) ViewPostHandler(w http.ResponseWriter, r *http.Request) {
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

	res, err := h.App.Staff.CoursePosts.ViewPost(ctx, &courseposts.ViewPostInput{UserID: userID, PostID: postID})
	if err != nil {
		switch {
		case errors.Is(err, courseposts.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	var mappedPosts []ViewPostPost
	if res.Posts != nil {
		mappedPosts = make([]ViewPostPost, 0, len(res.Posts))
		for _, item := range res.Posts {
			mappedPosts = append(mappedPosts, mapViewPostPost(item))
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access": res.Access,
		"posts":  mappedPosts,
	})
}
