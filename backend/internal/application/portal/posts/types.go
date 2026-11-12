package posts

import (
	"time"
)

type ViewPostPost struct {
	Attachments []ViewPostAttachment
	Author      ViewPostUser
	ID          string
	Title       string
	Body        string
	AccentColor string
	ShowUntil   *time.Time
	CreatedAt   time.Time
	EditedAt    *time.Time
}
type ViewPostAttachment struct {
	ID           string
	PresignedURL string
	Filename     string
	ContentType  string
}
type ViewPostUser struct {
	ID    string
	Name  string
	Email string
}
