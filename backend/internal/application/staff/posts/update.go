package posts

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"

	"app/internal/permissions"
	"app/internal/schoollogs"
)

type UpdatePostInput struct {
	InvalidPayload bool
	UserID         int64
	PostID         int64
	Data           Post
}

type UpdatePostResult struct{ AffectedRows int64 }

func (svc *Service) UpdatePost(ctx context.Context, i *UpdatePostInput) (*UpdatePostResult, error) {
	userID := i.UserID
	postID := i.PostID
	p := i.Data

	var schoolID int64
	if err := svc.DB.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM course_posts p
		JOIN courses c
		ON p.course_id = c.id
		JOIN grades g
		ON c.grade_id = g.id
		JOIN academic_years y
		ON g.academic_year_id = y.id
		WHERE p.id = $1
	`, postID).Scan(&schoolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}
	}

	if !permissions.Can(permissions.PermissionCoursePostUpdate, userID, schoolID, ctx, svc.DB) {
		return nil, ErrForbidden
	}

	if i.InvalidPayload {
		return nil, ErrBadRequest
	}

	validated, err := parseAndValidate(p)
	p = validated
	if err != nil {
		log.Println(err)
		return nil, ErrUnprocessableEntity
	}

	tx, err := svc.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var oldTitle string
	var oldBody string
	var oldAccentColor string
	if err := tx.QueryRowContext(ctx, `
		SELECT title, body, accent_color
		FROM course_posts
		WHERE id = $1
	`, postID).Scan(&oldTitle, &oldBody, &oldAccentColor); err != nil {
		log.Println(err)
		return nil, err
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE course_posts
		SET title = $1, body = $2, show_until = $3, accent_color = $4, edited_at = NOW()
		WHERE id = $5
	`, p.Title, p.Body, p.ShowUntil, p.AccentColor, postID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	details := "{user} updated the post title from '" + oldTitle + "' to '" + p.Title + "', changed accent color from '#" + oldAccentColor + "' to '#" + p.AccentColor + "', and changed the body length from " + strconv.Itoa(len(oldBody)) + " to " + strconv.Itoa(len(p.Body)) + " characters."
	if err := schoollogs.StoreSchoolLog(schoolID, userID, schoollogs.ActionPostEdit, schoollogs.TypeEdit, "Post updated", "{user} updated the post '"+p.Title+"'.", tx, ctx, svc.Sf, details); err != nil {
		log.Println(err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &UpdatePostResult{AffectedRows: affectedRows}, nil
}
