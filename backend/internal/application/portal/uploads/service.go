package uploads

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
	ErrForbidden           = errors.New("forbidden")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrBadRequest          = errors.New("bad request")
	ErrInternal            = errors.New("internal error")
)
