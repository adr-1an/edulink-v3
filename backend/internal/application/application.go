// Package application contains all service definitions
package application

import (
	portalPosts "app/internal/application/portal/posts"
	portalStudents "app/internal/application/portal/students"
	portalUploads "app/internal/application/portal/uploads"
	staffAssignments "app/internal/application/staff/assignments"
	staffCoursePostAttachments "app/internal/application/staff/coursepostattachments"
	staffCoursePosts "app/internal/application/staff/courseposts"
	staffCourses "app/internal/application/staff/courses"
	staffCourseStudents "app/internal/application/staff/coursestudents"
	staffGrades "app/internal/application/staff/grades"
	staffMembers "app/internal/application/staff/members"
	staffProfile "app/internal/application/staff/profile"
	staffRoles "app/internal/application/staff/roles"
	staffSubmissions "app/internal/application/staff/submissions"
	staffSubmissionScores "app/internal/application/staff/submissionscores"
	staffUploads "app/internal/application/staff/uploads"
	"database/sql"

	portalAuth "app/internal/application/portal/auth"
	portalProfile "app/internal/application/portal/profile"
	"app/internal/application/staff/academicyears"
	staffAuth "app/internal/application/staff/auth"
	"app/internal/application/staff/guardians"
	"app/internal/application/staff/logs"
	"app/internal/application/staff/portalusers"
	"app/internal/application/staff/schools"
	"app/internal/application/staff/students"

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
			Profile:               &staffProfile.Service{DB: db, Sf: sf, S3: s3},
			Uploads:               &staffUploads.Service{DB: db, Sf: sf, S3: s3},
			SubmissionScores:      &staffSubmissionScores.Service{DB: db, Sf: sf, S3: s3},
			Submissions:           &staffSubmissions.Service{DB: db, Sf: sf, S3: s3},
			Assignments:           &staffAssignments.Service{DB: db, Sf: sf, S3: s3},
			CourseStudents:        &staffCourseStudents.Service{DB: db, Sf: sf, S3: s3},
			CoursePostAttachments: &staffCoursePostAttachments.Service{DB: db, Sf: sf, S3: s3},
			CoursePosts:           &staffCoursePosts.Service{DB: db, Sf: sf, S3: s3},
			Courses:               &staffCourses.Service{DB: db, Sf: sf, S3: s3},
			Members:               &staffMembers.Service{DB: db, Sf: sf, S3: s3},
			Roles:                 &staffRoles.Service{DB: db, Sf: sf},
			Grades:                &staffGrades.Service{DB: db, Sf: sf},
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
	Profile               *staffProfile.Service
	Uploads               *staffUploads.Service
	SubmissionScores      *staffSubmissionScores.Service
	Submissions           *staffSubmissions.Service
	Assignments           *staffAssignments.Service
	CourseStudents        *staffCourseStudents.Service
	CoursePostAttachments *staffCoursePostAttachments.Service
	CoursePosts           *staffCoursePosts.Service
	Courses               *staffCourses.Service
	Members               *staffMembers.Service
	Roles                 *staffRoles.Service
	Grades                *staffGrades.Service
	AcademicYears         *academicyears.Service
	Auth                  *staffAuth.Service
	Schools               *schools.Service
	Students              *students.Service
	Guardians             *guardians.Service
	PortalUsers           *portalusers.Service
	Logs                  *logs.Service
}

type PortalServices struct {
	Students *portalStudents.Service
	Uploads  *portalUploads.Service
	Posts    *portalPosts.Service
	Auth     *portalAuth.Service
	Profile  *portalProfile.Service
}
