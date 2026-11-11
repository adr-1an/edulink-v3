// Package portal contains the main portal handlers
package portal

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
