// Package guardians contains all the main staff guardian services
package guardians

import (
	"database/sql"

	"github.com/sony/sonyflake/v2"
)

type Service struct {
	DB *sql.DB
	Sf *sonyflake.Sonyflake
}
