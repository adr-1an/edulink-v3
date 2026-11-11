// Package auth contains all the main staff auth services
package auth

import (
	"database/sql"

	"github.com/sony/sonyflake/v2"
)

type Service struct {
	DB *sql.DB
	Sf *sonyflake.Sonyflake
}
