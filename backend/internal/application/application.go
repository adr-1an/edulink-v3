// Package application contains all service definitions
package application

import (
	"database/sql"

	portalAuth "app/internal/application/portal/auth"
	portalPosts "app/internal/application/portal/posts"
	portalProfile "app/internal/application/portal/profile"
	portalStudents "app/internal/application/portal/students"
	portalUploads "app/internal/application/portal/uploads"
	"app/internal/application/staff/academicyears"
	staffAssignments "app/internal/application/staff/assignments"
	staffAttachments "app/internal/application/staff/attachments"
	staffAuth "app/internal/application/staff/auth"
	staffCourses "app/internal/application/staff/courses"
	staffEnrollments "app/internal/application/staff/enrollments"
	staffGrades "app/internal/application/staff/grades"
	"app/internal/application/staff/guardians"
	"app/internal/application/staff/logs"
	staffMembers "app/internal/application/staff/members"
	"app/internal/application/staff/portalusers"
	staffPosts "app/internal/application/staff/posts"
	staffProfile "app/internal/application/staff/profile"
	staffRoles "app/internal/application/staff/roles"
	"app/internal/application/staff/schools"
	staffScores "app/internal/application/staff/scores"
	"app/internal/application/staff/students"
	staffSubmissions "app/internal/application/staff/submissions"
	staffUploads "app/internal/application/staff/uploads"

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
			Profile:     &staffProfile.Service{DB: db, Sf: sf, S3: s3},
			Uploads:     &staffUploads.Service{DB: db, Sf: sf, S3: s3},
			Scores:      &staffScores.Service{DB: db, Sf: sf, S3: s3},
			Submissions: &staffSubmissions.Service{DB: db, Sf: sf, S3: s3},
			Assignments: &staffAssignments.Service{DB: db, Sf: sf, S3: s3},
			Enrollments: &staffEnrollments.Service{DB: db, Sf: sf, S3: s3},
			Attachments: &staffAttachments.Service{DB: db, Sf: sf, S3: s3},
			Posts:       &staffPosts.Service{DB: db, Sf: sf, S3: s3},
			Courses:     &staffCourses.Service{DB: db, Sf: sf, S3: s3},
			Members:     &staffMembers.Service{DB: db, Sf: sf, S3: s3},
			Roles:       &staffRoles.Service{DB: db, Sf: sf},
			Grades:      &staffGrades.Service{DB: db, Sf: sf},
			AcademicYears: &academicyears.Service{
				DB: db,
				Sf: sf,
			},
			Auth: &staffAuth.Service{
				Schools: &schools.Service{DB: db, Sf: sf},
				DB:      db,
				Sf:      sf,
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
			Guardians: &guardians.Service{
				DB: db,
				Sf: sf,
			},
			PortalUsers: &portalusers.Service{
				DB: db,
				Sf: sf,
			},
			Logs: &logs.Service{
				DB: db,
				Sf: sf,
			},
		},
		Portal: PortalServices{
			Students: &portalStudents.Service{DB: db, S3: s3, Sf: sf},
			Uploads:  &portalUploads.Service{DB: db, S3: s3},
			Posts:    &portalPosts.Service{DB: db, S3: s3},
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
	Profile       *staffProfile.Service
	Uploads       *staffUploads.Service
	Scores        *staffScores.Service
	Submissions   *staffSubmissions.Service
	Assignments   *staffAssignments.Service
	Enrollments   *staffEnrollments.Service
	Attachments   *staffAttachments.Service
	Posts         *staffPosts.Service
	Courses       *staffCourses.Service
	Members       *staffMembers.Service
	Roles         *staffRoles.Service
	Grades        *staffGrades.Service
	AcademicYears *academicyears.Service
	Auth          *staffAuth.Service
	Schools       *schools.Service
	Students      *students.Service
	Guardians     *guardians.Service
	PortalUsers   *portalusers.Service
	Logs          *logs.Service
}

type PortalServices struct {
	Students *portalStudents.Service
	Uploads  *portalUploads.Service
	Posts    *portalPosts.Service
	Auth     *portalAuth.Service
	Profile  *portalProfile.Service
}
