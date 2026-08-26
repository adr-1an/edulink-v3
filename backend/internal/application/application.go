package application

import (
	portalAuth "app/internal/application/portal/auth"
	staffAuth "app/internal/application/staff/auth"
	"app/internal/application/staff/schools"
	"database/sql"

	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

func New(
	db *sql.DB,
	sf *sonyflake.Sonyflake,
	s3 *minio.Client,
) *Application {
	return &Application{
		Staff: StaffServices{
			Auth: &staffAuth.Service{
				DB: db,
				Sf: sf,
			},
			Schools: &schools.Service{
				DB: db,
				Sf: sf,
			},
		},
		Portal: PortalServices{
			Auth: &portalAuth.Service{
				DB: db,
				Sf: sf,
			},
		},
	}
}

type Application struct {
	Staff  StaffServices
	Portal PortalServices
}

type StaffServices struct {
	Auth    *staffAuth.Service
	Schools *schools.Service
}

type PortalServices struct {
	Auth *portalAuth.Service
}
