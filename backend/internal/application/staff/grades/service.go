// Package grades contains the main staff grades services
package grades

import (
	"database/sql"

	"github.com/sony/sonyflake/v2"
)

type Service struct {
	DB *sql.DB
	Sf *sonyflake.Sonyflake
}
