// Package members contains the main staff members services
package members

import (
	"database/sql"

	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

type Service struct {
	DB *sql.DB
	Sf *sonyflake.Sonyflake
	S3 *minio.Client
}
