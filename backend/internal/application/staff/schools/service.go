// Package schools contains serices for schools
package schools

import (
	"database/sql"

	"github.com/sony/sonyflake/v2"
)

type Service struct {
	DB *sql.DB
	Sf *sonyflake.Sonyflake
}
