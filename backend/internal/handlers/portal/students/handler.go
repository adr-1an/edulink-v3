package students

import (
	"app/internal/application"
	"database/sql"
)

type Handler struct {
	DB  *sql.DB
	App *application.Application
}
