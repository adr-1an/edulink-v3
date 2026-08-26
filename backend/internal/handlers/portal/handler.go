package portal

import (
	"app/internal/application"
	"database/sql"

	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

type Handler struct {
	DB  *sql.DB
	Sf  *sonyflake.Sonyflake
	S3  *minio.Client
	App *application.Application
}
