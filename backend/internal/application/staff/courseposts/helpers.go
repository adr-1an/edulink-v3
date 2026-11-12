package courseposts

import (
	"errors"
	"strings"
	"time"
)

type Post struct {
	Title       string
	Body        string
	AccentColor string
	ShowUntil   *time.Time
}

func parseAndValidate(p Post) (Post, error) {
	// Parse
	p.Title = strings.TrimSpace(p.Title)
	p.Body = strings.TrimSpace(p.Body)
	p.AccentColor = strings.TrimSpace(strings.TrimPrefix(p.AccentColor, "#"))

	// Validate
	if p.Title == "" || p.Body == "" || p.AccentColor == "" {
		return Post{}, errors.New("missing fields")
	}

	if len(p.Title) > 32 || len(p.Body) > 2048 || len(p.AccentColor) != 6 {
		return Post{}, errors.New("invalid lengths")
	}

	return p, nil
}
