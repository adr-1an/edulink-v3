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

type DeletePostInput struct {
	UserID int64
	PostID int64
}

type DeletePostResult struct{ AffectedRows int64 }

func (svc *Service) DeletePost(ctx context.Context, i *DeletePostInput) (*DeletePostResult, error) {
	userID := i.UserID
	postID := i.PostID

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

	if !permissions.Can(permissions.PermissionCoursePostDelete, userID, schoolID, ctx, svc.DB) {
		return nil, ErrForbidden
	}

	tx, err := svc.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var postTitle string
	if err := tx.QueryRowContext(ctx, `
		SELECT title FROM course_posts WHERE id = $1
	`, postID).Scan(&postTitle); err != nil {
		log.Println(err)
		return nil, err
	}

	res, err := tx.ExecContext(ctx, `
		DELETE FROM course_posts WHERE id = $1
	`, postID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if err := schoollogs.StoreSchoolLog(schoolID, userID, schoollogs.ActionPostDelete, schoollogs.TypeDelete, "Post deleted", "{user} deleted the post '"+postTitle+"'.", tx, ctx, svc.Sf, "{user} deleted post '"+postTitle+"' with ID "+strconv.FormatInt(postID, 10)+"."); err != nil {
		log.Println(err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &DeletePostResult{AffectedRows: affectedRows}, nil
}
