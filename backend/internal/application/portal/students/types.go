package students

import (
	"app/internal/submissions"
	"time"
)

type ListCoursesCourse struct {
	ID          string
	Grade       ListCoursesGrade
	Name        string
	Description string
	AccentColor string
}
type ListCoursesGrade struct {
	ID   string
	Name string
}
type CourseDashboardCourse struct {
	Grade       CourseDashboardGrade
	ID          string
	Name        string
	Description string
	AccentColor string
}
type CourseDashboardGrade struct {
	ID    string
	Name  string
	Level int
}
type CourseDashboardAssignment struct {
	ReferencedPost     *CourseDashboardReferencedPost
	Submission         *AssignmentSubmission
	ID                 string
	Title              string
	Description        string
	DueDate            *time.Time
	SubmissionsEnabled bool
	SubmissionsCloseAt *time.Time
	CreatedAt          time.Time
}
type CourseDashboardReferencedPost struct {
	ID    string
	Title string
}
type CourseDashboardPost struct {
	ID          string
	Attachments []CourseDashboardAttachment
	Author      CourseDashboardUser
	Title       string
	Body        string
	AccentColor string
	CreatedAt   time.Time
	EditedAt    *time.Time
}
type CourseDashboardAttachment struct {
	ID           string
	PresignedURL string
	Name         string
	ContentType  string
}
type CourseDashboardUser struct {
	ID     string
	Name   string
	Email  string
	PfpURL *string
}
type ListAllAssignmentsAssignment struct {
	RefPost            *ListAllAssignmentsRefPost
	Course             ListAllAssignmentsCourse
	Submission         *AssignmentSubmission
	ID                 string
	Title              string
	Description        string
	DueDate            *time.Time
	SubmissionsEnabled bool
	SubmissionsCloseAt *time.Time
	CreatedAt          time.Time
}
type ListAllAssignmentsCourse struct {
	ID   string
	Name string
}
type ListAllAssignmentsRefPost struct {
	ID    string
	Title string
	Body  string
	Color string
}
type AssignmentGradeUser struct {
	ID    string
	Name  string
	Email string
	Phone *string
}
type SubmissionGrade struct {
	Score    int
	Notes    *string
	GradedAt time.Time
	GradedBy AssignmentGradeUser
}
type AssignmentSubmission struct {
	ID          string
	Status      submissions.SubmissionStatus
	Grade       *SubmissionGrade
	Notes       string
	Attachments []AssignmentSubmissionAttachment
	CreatedAt   time.Time
}
type AssignmentSubmissionAttachment struct {
	ID           string
	FileName     string
	FileSize     int64
	ContentType  string
	PresignedURL string
}
