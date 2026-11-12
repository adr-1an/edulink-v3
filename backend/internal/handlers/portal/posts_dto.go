package portal

import (
	svc "app/internal/application/portal/posts"
	"time"
)

type ViewPostPost struct {
	Attachments []ViewPostAttachment `json:"attachments"`
	Author      ViewPostUser         `json:"author"`
	ID          string               `json:"id"`
	Title       string               `json:"title"`
	Body        string               `json:"body"`
	AccentColor string               `json:"accentColor"`
	ShowUntil   *time.Time           `json:"showUntil"`
	CreatedAt   time.Time            `json:"createdAt"`
	EditedAt    *time.Time           `json:"editedAt"`
}

func mapViewPostPost(v svc.ViewPostPost) ViewPostPost {
	return ViewPostPost{
		Attachments: mapViewPostAttachmentSlice(v.Attachments),
		Author:      mapViewPostUser(v.Author),
		ID:          v.ID,
		Title:       v.Title,
		Body:        v.Body,
		AccentColor: v.AccentColor,
		ShowUntil:   v.ShowUntil,
		CreatedAt:   v.CreatedAt,
		EditedAt:    v.EditedAt,
	}
}

type ViewPostAttachment struct {
	ID           string `json:"id"`
	PresignedURL string `json:"presignedUrl"`
	Filename     string `json:"fileName"`
	ContentType  string `json:"contentType"`
}

func mapViewPostAttachment(v svc.ViewPostAttachment) ViewPostAttachment {
	return ViewPostAttachment{
		ID:           v.ID,
		PresignedURL: v.PresignedURL,
		Filename:     v.Filename,
		ContentType:  v.ContentType,
	}
}

func mapViewPostAttachmentSlice(v []svc.ViewPostAttachment) []ViewPostAttachment {
	if v == nil {
		return nil
	}
	r := make([]ViewPostAttachment, len(v))
	for k := range v {
		r[k] = mapViewPostAttachment(v[k])
	}
	return r
}

type ViewPostUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func mapViewPostUser(v svc.ViewPostUser) ViewPostUser {
	return ViewPostUser{
		ID:    v.ID,
		Name:  v.Name,
		Email: v.Email,
	}
}
