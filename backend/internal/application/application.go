package application

import (
	portalAuth "app/internal/application/portal/auth"
	portalProfile "app/internal/application/portal/profile"
	staffAuth "app/internal/application/staff/auth"
	"app/internal/application/staff/schools"
	"app/internal/application/staff/students"
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
			Students: &students.Service{
				DB: db,
				Sf: sf,
				S3: s3,
			},
		},
		Portal: PortalServices{
			Auth: &portalAuth.Service{
				DB: db,
				Sf: sf,
			},
			Profile: &portalProfile.Service{
				DB: db,
				Sf: sf,
				S3: s3,
			},
		},
	}
}

type Application struct {
	Staff  StaffServices
	Portal PortalServices
}

type StaffServices struct {
	Auth     *staffAuth.Service
	Schools  *schools.Service
	Students *students.Service
}

type PortalServices struct {
	Auth    *portalAuth.Service
	Profile *portalProfile.Service
}
