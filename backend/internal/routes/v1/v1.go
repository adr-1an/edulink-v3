package v1

import (
	"app/internal/routes/v1/portal"
	"app/internal/routes/v1/staff"
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
) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	// API v1
	r.Route("/v1", func(r chi.Router) {
		// Staff app
		r.Mount("/staff", staff.MainStaffRoutes(db, sf, s3))

		// Portal app
		r.Mount("/portal", portal.MainPortalRoutes(db, sf, s3))
	})

	return r
}
