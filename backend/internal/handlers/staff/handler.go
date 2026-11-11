// Package staff contains the main staff handlers
package staff

import (
	"database/sql"

	"app/internal/application"

	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

type Handler struct {
	DB  *sql.DB
	Sf  *sonyflake.Sonyflake
	S3  *minio.Client
	App *application.Application
}

// Once all app logic for all handlers is migrated to services, S3 and Sf should be removed from the struct.
