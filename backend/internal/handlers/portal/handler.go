// Package portal contains the main portal handlers
package portal

import (
	"database/sql"

	"app/internal/application"
)

type Handler struct {
	DB  *sql.DB
	App *application.Application
}
