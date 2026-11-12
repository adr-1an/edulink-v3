package students

import (
	"database/sql"
	"errors"

	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

type Service struct {
	DB *sql.DB
	Sf *sonyflake.Sonyflake
	S3 *minio.Client
}

var (
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrBadRequest          = errors.New("bad request")
	ErrInternal            = errors.New("internal error")
)
