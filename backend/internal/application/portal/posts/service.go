package posts

import (
	"database/sql"
	"errors"

	"github.com/minio/minio-go/v7"
)

type Service struct {
	DB *sql.DB
	S3 *minio.Client
}

var (
	ErrForbidden = errors.New("forbidden")
)
