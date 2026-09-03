package v1

import (
	"app/internal/application"
	portalHandlers "app/internal/handlers/portal"
	"app/internal/handlers/portal/students"
	"app/internal/handlers/staff"
	"app/internal/routes/v1/portal"
	staffRoutes "app/internal/routes/v1/staff"
	"database/sql"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

func MainRouter(
	db *sql.DB,
	sf *sonyflake.Sonyflake,
	s3 *minio.Client,
	app *application.Application,
	trustForwarded bool,
) chi.Router {
	r := chi.NewRouter()

	// Staff handler
	h := staff.Handler{
		DB:  db,
		Sf:  sf,
		S3:  s3,
		App: app,
	}

	// Portal handlers
	portalHandler := portalHandlers.Handler{
		DB:  db,
		Sf:  sf,
		S3:  s3,
		App: app,
	}
	studentHandler := students.Handler{
		DB:  db,
		Sf:  sf,
		S3:  s3,
		App: app,
	}

	r.Use(middleware.RequestID)

	if trustForwarded {
		r.Use(middleware.ClientIPFromHeader("CF-Connecting-IP"))
	} else {
		r.Use(middleware.ClientIPFromRemoteAddr)
	}

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	// API v1
	r.Route("/v1", func(r chi.Router) {
		// Staff app
		r.Mount("/staff", staffRoutes.MainStaffRoutes(&h))

		// Portal app
		r.Mount("/portal", portal.MainPortalRoutes(&portalHandler, &studentHandler))
	})

	return r
}
