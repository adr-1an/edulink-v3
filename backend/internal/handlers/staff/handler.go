// Package staff contains the main staff handlers
package staff

import (
	"database/sql"

	"app/internal/application"
)

type Handler struct {
	DB  *sql.DB
	App *application.Application
}
