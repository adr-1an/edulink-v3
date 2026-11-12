package courseposts

import (
	"context"
	"log"

	"app/internal/permissions"
	"app/internal/schoollogs"
)

type CreatePostInput struct {
	InvalidPayload bool
	UserID         int64
	CourseID       int64
	Data           Post
}

type CreatePostResult struct {
	ID           int64
	AffectedRows int64
}

func (svc *Service) CreatePost(ctx context.Context, i *CreatePostInput) (*CreatePostResult, error) {
	userID := i.UserID
	courseID := i.CourseID
	p := i.Data

	var schoolID int64
	if err := svc.DB.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM courses c
		JOIN grades g
		ON c.grade_id = g.id
		JOIN academic_years y
		ON G.academic_year_id = y.id
		WHERE c.id = $1
	`, courseID).Scan(&schoolID); err != nil {
		log.Println(err)
		return nil, err
	}

	if !permissions.Can(permissions.PermissionCoursePostCreate, userID, schoolID, ctx, svc.DB) {
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

	id, err := svc.Sf.NextID()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tx, err := svc.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO course_posts (id, course_id, author_id, title, body, accent_color, show_until)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, courseID, userID, p.Title, p.Body, p.AccentColor, p.ShowUntil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if err := schoollogs.StoreSchoolLog(schoolID, userID, schoollogs.ActionPostCreate, schoollogs.TypeCreate, "Post created", "{user} created the post '"+p.Title+"'.", tx, ctx, svc.Sf, "{user} created post '"+p.Title+"' with accent color '#"+p.AccentColor+"'."); err != nil {
		log.Println(err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &CreatePostResult{ID: id, AffectedRows: affectedRows}, nil
}
