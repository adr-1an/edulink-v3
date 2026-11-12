// Package academicyears contains the main staff academic year services
package academicyears

import (
	"database/sql"

	"github.com/sony/sonyflake/v2"
)

type Service struct {
	DB *sql.DB
	Sf *sonyflake.Sonyflake
}
