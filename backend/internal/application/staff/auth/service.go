// Package auth contains all the main staff auth services
package auth

import (
	"database/sql"

	"app/internal/application/staff/schools"

	"github.com/sony/sonyflake/v2"
)

type Service struct {
	Schools *schools.Service
	DB      *sql.DB
	Sf      *sonyflake.Sonyflake
}
