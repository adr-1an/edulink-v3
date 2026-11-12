package logs

import (
	"context"
	"log"

	"app/internal/permissions"
)

type ListLogsInput struct {
	SchoolID int64
	UserID   int64
}

type User struct {
	ID    int64
	Name  string
	Email string
}

type SchoolLog struct {
	ID        int64
	User      User
	Action    string
	Type      string
	Title     string
	Message   string
	Details   string
	CreatedAt string
}

type ListLogsResult struct {
	Access   permissions.AccessData
	Logs     []SchoolLog
	SchoolID int64
}

func (s *Service) ListLogs(ctx context.Context, i *ListLogsInput) (*ListLogsResult, error) {
	if !permissions.Can(permissions.PermissionLogsList, i.UserID, i.SchoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	rows, err := s.DB.QueryContext(ctx, `
		SELECT
		    l.id,
		    l.action,
		    l.type,
		    l.title,
		    l.message,
		    l.details,
		    l.created_at,
		    u.id,
		    u.name,
		    u.email
		FROM school_logs l
		JOIN users u ON u.id = l.by_user
		WHERE school_id = $1
		AND l.created_at > NOW() - INTERVAL '30 days'
	`, i.SchoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	logs := make([]SchoolLog, 0)
	for rows.Next() {
		var l SchoolLog

		if err := rows.Scan(
			&l.ID,
			&l.Action,
			&l.Type,
			&l.Title,
			&l.Message,
			&l.Details,
			&l.CreatedAt,
			&l.User.ID,
			&l.User.Name,
			&l.User.Email,
		); err != nil {
			log.Println(err)
			return nil, err
		}

		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	access, err := permissions.GetAllUserPermissions(ctx, s.DB, i.UserID, i.SchoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &ListLogsResult{
		Access:   access,
		SchoolID: i.SchoolID,
		Logs:     logs,
	}, nil
}
