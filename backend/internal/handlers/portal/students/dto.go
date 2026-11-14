package students

import (
	"time"

	svc "app/internal/application/portal/students"
	"app/internal/submissions"
)

type ListCoursesCourse struct {
	ID          string           `json:"id"`
	Grade       ListCoursesGrade `json:"grade"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	AccentColor string           `json:"accentColor"`
}

func mapListCoursesCourse(v svc.ListCoursesCourse) ListCoursesCourse {
	return ListCoursesCourse{
		ID:          v.ID,
		Grade:       mapListCoursesGrade(v.Grade),
		Name:        v.Name,
		Description: v.Description,
		AccentColor: v.AccentColor,
	}
}

func mapListCoursesCoursePtr(v *svc.ListCoursesCourse) *ListCoursesCourse {
	if v == nil {
		return nil
	}
	r := mapListCoursesCourse(*v)
	return &r
}

func mapListCoursesCourseSlice(v []svc.ListCoursesCourse) []ListCoursesCourse {
	if v == nil {
		return nil
	}
	r := make([]ListCoursesCourse, len(v))
	for k := range v {
		r[k] = mapListCoursesCourse(v[k])
	}
	return r
}

type ListCoursesGrade struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func mapListCoursesGrade(v svc.ListCoursesGrade) ListCoursesGrade {
	return ListCoursesGrade{
		ID:   v.ID,
		Name: v.Name,
	}
}

func mapListCoursesGradePtr(v *svc.ListCoursesGrade) *ListCoursesGrade {
	if v == nil {
		return nil
	}
	r := mapListCoursesGrade(*v)
	return &r
}

func mapListCoursesGradeSlice(v []svc.ListCoursesGrade) []ListCoursesGrade {
	if v == nil {
		return nil
	}
	r := make([]ListCoursesGrade, len(v))
	for k := range v {
		r[k] = mapListCoursesGrade(v[k])
	}
	return r
}

type CourseDashboardCourse struct {
	Grade       CourseDashboardGrade `json:"grade"`
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	AccentColor string               `json:"accentColor"`
}

func mapCourseDashboardCourse(v svc.CourseDashboardCourse) CourseDashboardCourse {
	return CourseDashboardCourse{
		Grade:       mapCourseDashboardGrade(v.Grade),
		ID:          v.ID,
		Name:        v.Name,
		Description: v.Description,
		AccentColor: v.AccentColor,
	}
}

func mapCourseDashboardCoursePtr(v *svc.CourseDashboardCourse) *CourseDashboardCourse {
	if v == nil {
		return nil
	}
	r := mapCourseDashboardCourse(*v)
	return &r
}

func mapCourseDashboardCourseSlice(v []svc.CourseDashboardCourse) []CourseDashboardCourse {
	if v == nil {
		return nil
	}
	r := make([]CourseDashboardCourse, len(v))
	for k := range v {
		r[k] = mapCourseDashboardCourse(v[k])
	}
	return r
}

type CourseDashboardGrade struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Level int    `json:"level"`
}

func mapCourseDashboardGrade(v svc.CourseDashboardGrade) CourseDashboardGrade {
	return CourseDashboardGrade{
		ID:    v.ID,
		Name:  v.Name,
		Level: v.Level,
	}
}

func mapCourseDashboardGradePtr(v *svc.CourseDashboardGrade) *CourseDashboardGrade {
	if v == nil {
		return nil
	}
	r := mapCourseDashboardGrade(*v)
	return &r
}

func mapCourseDashboardGradeSlice(v []svc.CourseDashboardGrade) []CourseDashboardGrade {
	if v == nil {
		return nil
	}
	r := make([]CourseDashboardGrade, len(v))
	for k := range v {
		r[k] = mapCourseDashboardGrade(v[k])
	}
	return r
}

type CourseDashboardAssignment struct {
	ReferencedPost     *CourseDashboardReferencedPost `json:"referencedPost"`
	Submission         *AssignmentSubmission          `json:"submission"`
	ID                 string                         `json:"id"`
	Title              string                         `json:"title"`
	Description        string                         `json:"description"`
	DueDate            *time.Time                     `json:"dueDate"`
	SubmissionsEnabled bool                           `json:"submissionsEnabled"`
	SubmissionsCloseAt *time.Time                     `json:"submissionsCloseAt"`
	CreatedAt          time.Time                      `json:"createdAt"`
}

func mapCourseDashboardAssignment(v svc.CourseDashboardAssignment) CourseDashboardAssignment {
	return CourseDashboardAssignment{
		ReferencedPost:     mapCourseDashboardReferencedPostPtr(v.ReferencedPost),
		Submission:         mapAssignmentSubmissionPtr(v.Submission),
		ID:                 v.ID,
		Title:              v.Title,
		Description:        v.Description,
		DueDate:            v.DueDate,
		SubmissionsEnabled: v.SubmissionsEnabled,
		SubmissionsCloseAt: v.SubmissionsCloseAt,
		CreatedAt:          v.CreatedAt,
	}
}

func mapCourseDashboardAssignmentPtr(v *svc.CourseDashboardAssignment) *CourseDashboardAssignment {
	if v == nil {
		return nil
	}
	r := mapCourseDashboardAssignment(*v)
	return &r
}

func mapCourseDashboardAssignmentSlice(v []svc.CourseDashboardAssignment) []CourseDashboardAssignment {
	if v == nil {
		return nil
	}
	r := make([]CourseDashboardAssignment, len(v))
	for k := range v {
		r[k] = mapCourseDashboardAssignment(v[k])
	}
	return r
}

type CourseDashboardReferencedPost struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func mapCourseDashboardReferencedPost(v svc.CourseDashboardReferencedPost) CourseDashboardReferencedPost {
	return CourseDashboardReferencedPost{
		ID:    v.ID,
		Title: v.Title,
	}
}

func mapCourseDashboardReferencedPostPtr(v *svc.CourseDashboardReferencedPost) *CourseDashboardReferencedPost {
	if v == nil {
		return nil
	}
	r := mapCourseDashboardReferencedPost(*v)
	return &r
}

func mapCourseDashboardReferencedPostSlice(v []svc.CourseDashboardReferencedPost) []CourseDashboardReferencedPost {
	if v == nil {
		return nil
	}
	r := make([]CourseDashboardReferencedPost, len(v))
	for k := range v {
		r[k] = mapCourseDashboardReferencedPost(v[k])
	}
	return r
}

type CourseDashboardPost struct {
	ID          string                      `json:"id"`
	Attachments []CourseDashboardAttachment `json:"attachments"`
	Author      CourseDashboardUser         `json:"author"`
	Title       string                      `json:"title"`
	Body        string                      `json:"body"`
	AccentColor string                      `json:"accentColor"`
	CreatedAt   time.Time                   `json:"createdAt"`
	EditedAt    *time.Time                  `json:"editedAt"`
}

func mapCourseDashboardPost(v svc.CourseDashboardPost) CourseDashboardPost {
	return CourseDashboardPost{
		ID:          v.ID,
		Attachments: mapCourseDashboardAttachmentSlice(v.Attachments),
		Author:      mapCourseDashboardUser(v.Author),
		Title:       v.Title,
		Body:        v.Body,
		AccentColor: v.AccentColor,
		CreatedAt:   v.CreatedAt,
		EditedAt:    v.EditedAt,
	}
}

func mapCourseDashboardPostPtr(v *svc.CourseDashboardPost) *CourseDashboardPost {
	if v == nil {
		return nil
	}
	r := mapCourseDashboardPost(*v)
	return &r
}

func mapCourseDashboardPostSlice(v []svc.CourseDashboardPost) []CourseDashboardPost {
	if v == nil {
		return nil
	}
	r := make([]CourseDashboardPost, len(v))
	for k := range v {
		r[k] = mapCourseDashboardPost(v[k])
	}
	return r
}

type CourseDashboardAttachment struct {
	ID           string `json:"id"`
	PresignedURL string `json:"presignedUrl"`
	Name         string `json:"fileName"`
	ContentType  string `json:"contentType"`
}

func mapCourseDashboardAttachment(v svc.CourseDashboardAttachment) CourseDashboardAttachment {
	return CourseDashboardAttachment{
		ID:           v.ID,
		PresignedURL: v.PresignedURL,
		Name:         v.Name,
		ContentType:  v.ContentType,
	}
}

func mapCourseDashboardAttachmentPtr(v *svc.CourseDashboardAttachment) *CourseDashboardAttachment {
	if v == nil {
		return nil
	}
	r := mapCourseDashboardAttachment(*v)
	return &r
}

func mapCourseDashboardAttachmentSlice(v []svc.CourseDashboardAttachment) []CourseDashboardAttachment {
	if v == nil {
		return nil
	}
	r := make([]CourseDashboardAttachment, len(v))
	for k := range v {
		r[k] = mapCourseDashboardAttachment(v[k])
	}
	return r
}

type CourseDashboardUser struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Email  string  `json:"email"`
	PfpURL *string `json:"profilePictureURL"`
}

func mapCourseDashboardUser(v svc.CourseDashboardUser) CourseDashboardUser {
	return CourseDashboardUser{
		ID:     v.ID,
		Name:   v.Name,
		Email:  v.Email,
		PfpURL: v.PfpURL,
	}
}

func mapCourseDashboardUserPtr(v *svc.CourseDashboardUser) *CourseDashboardUser {
	if v == nil {
		return nil
	}
	r := mapCourseDashboardUser(*v)
	return &r
}

func mapCourseDashboardUserSlice(v []svc.CourseDashboardUser) []CourseDashboardUser {
	if v == nil {
		return nil
	}
	r := make([]CourseDashboardUser, len(v))
	for k := range v {
		r[k] = mapCourseDashboardUser(v[k])
	}
	return r
}

type ListAllAssignmentsAssignment struct {
	RefPost            *ListAllAssignmentsRefPost `json:"referencedPost"`
	Course             ListAllAssignmentsCourse   `json:"course"`
	Submission         *AssignmentSubmission      `json:"submission"`
	ID                 string                     `json:"id"`
	Title              string                     `json:"title"`
	Description        string                     `json:"description"`
	DueDate            *time.Time                 `json:"dueDate"`
	SubmissionsEnabled bool                       `json:"submissionsEnabled"`
	SubmissionsCloseAt *time.Time                 `json:"submissionsCloseAt"`
	CreatedAt          time.Time                  `json:"createdAt"`
}

func mapListAllAssignmentsAssignment(v svc.ListAllAssignmentsAssignment) ListAllAssignmentsAssignment {
	return ListAllAssignmentsAssignment{
		RefPost:            mapListAllAssignmentsRefPostPtr(v.RefPost),
		Course:             mapListAllAssignmentsCourse(v.Course),
		Submission:         mapAssignmentSubmissionPtr(v.Submission),
		ID:                 v.ID,
		Title:              v.Title,
		Description:        v.Description,
		DueDate:            v.DueDate,
		SubmissionsEnabled: v.SubmissionsEnabled,
		SubmissionsCloseAt: v.SubmissionsCloseAt,
		CreatedAt:          v.CreatedAt,
	}
}

func mapListAllAssignmentsAssignmentPtr(v *svc.ListAllAssignmentsAssignment) *ListAllAssignmentsAssignment {
	if v == nil {
		return nil
	}
	r := mapListAllAssignmentsAssignment(*v)
	return &r
}

func mapListAllAssignmentsAssignmentSlice(v []svc.ListAllAssignmentsAssignment) []ListAllAssignmentsAssignment {
	if v == nil {
		return nil
	}
	r := make([]ListAllAssignmentsAssignment, len(v))
	for k := range v {
		r[k] = mapListAllAssignmentsAssignment(v[k])
	}
	return r
}

type ListAllAssignmentsCourse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func mapListAllAssignmentsCourse(v svc.ListAllAssignmentsCourse) ListAllAssignmentsCourse {
	return ListAllAssignmentsCourse{
		ID:   v.ID,
		Name: v.Name,
	}
}

func mapListAllAssignmentsCoursePtr(v *svc.ListAllAssignmentsCourse) *ListAllAssignmentsCourse {
	if v == nil {
		return nil
	}
	r := mapListAllAssignmentsCourse(*v)
	return &r
}

func mapListAllAssignmentsCourseSlice(v []svc.ListAllAssignmentsCourse) []ListAllAssignmentsCourse {
	if v == nil {
		return nil
	}
	r := make([]ListAllAssignmentsCourse, len(v))
	for k := range v {
		r[k] = mapListAllAssignmentsCourse(v[k])
	}
	return r
}

type ListAllAssignmentsRefPost struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Color string `json:"accentColor"`
}

func mapListAllAssignmentsRefPost(v svc.ListAllAssignmentsRefPost) ListAllAssignmentsRefPost {
	return ListAllAssignmentsRefPost{
		ID:    v.ID,
		Title: v.Title,
		Body:  v.Body,
		Color: v.Color,
	}
}

func mapListAllAssignmentsRefPostPtr(v *svc.ListAllAssignmentsRefPost) *ListAllAssignmentsRefPost {
	if v == nil {
		return nil
	}
	r := mapListAllAssignmentsRefPost(*v)
	return &r
}

func mapListAllAssignmentsRefPostSlice(v []svc.ListAllAssignmentsRefPost) []ListAllAssignmentsRefPost {
	if v == nil {
		return nil
	}
	r := make([]ListAllAssignmentsRefPost, len(v))
	for k := range v {
		r[k] = mapListAllAssignmentsRefPost(v[k])
	}
	return r
}

type AssignmentGradeUser struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Email string  `json:"email"`
	Phone *string `json:"phone"`
}

func mapAssignmentGradeUser(v svc.AssignmentGradeUser) AssignmentGradeUser {
	return AssignmentGradeUser{
		ID:    v.ID,
		Name:  v.Name,
		Email: v.Email,
		Phone: v.Phone,
	}
}

func mapAssignmentGradeUserPtr(v *svc.AssignmentGradeUser) *AssignmentGradeUser {
	if v == nil {
		return nil
	}
	r := mapAssignmentGradeUser(*v)
	return &r
}

func mapAssignmentGradeUserSlice(v []svc.AssignmentGradeUser) []AssignmentGradeUser {
	if v == nil {
		return nil
	}
	r := make([]AssignmentGradeUser, len(v))
	for k := range v {
		r[k] = mapAssignmentGradeUser(v[k])
	}
	return r
}

type SubmissionGrade struct {
	Score    int                 `json:"score"`
	Notes    *string             `json:"notes"`
	GradedAt time.Time           `json:"gradedAt"`
	GradedBy AssignmentGradeUser `json:"gradedBy"`
}

func mapSubmissionGrade(v svc.SubmissionGrade) SubmissionGrade {
	return SubmissionGrade{
		Score:    v.Score,
		Notes:    v.Notes,
		GradedAt: v.GradedAt,
		GradedBy: mapAssignmentGradeUser(v.GradedBy),
	}
}

func mapSubmissionGradePtr(v *svc.SubmissionGrade) *SubmissionGrade {
	if v == nil {
		return nil
	}
	r := mapSubmissionGrade(*v)
	return &r
}

func mapSubmissionGradeSlice(v []svc.SubmissionGrade) []SubmissionGrade {
	if v == nil {
		return nil
	}
	r := make([]SubmissionGrade, len(v))
	for k := range v {
		r[k] = mapSubmissionGrade(v[k])
	}
	return r
}

type AssignmentSubmission struct {
	ID          string                           `json:"id"`
	Status      submissions.SubmissionStatus     `json:"status"`
	Grade       *SubmissionGrade                 `json:"grade"`
	Notes       string                           `json:"notes"`
	Attachments []AssignmentSubmissionAttachment `json:"attachments"`
	CreatedAt   time.Time                        `json:"createdAt"`
}

func mapAssignmentSubmission(v svc.AssignmentSubmission) AssignmentSubmission {
	return AssignmentSubmission{
		ID:          v.ID,
		Status:      v.Status,
		Grade:       mapSubmissionGradePtr(v.Grade),
		Notes:       v.Notes,
		Attachments: mapAssignmentSubmissionAttachmentSlice(v.Attachments),
		CreatedAt:   v.CreatedAt,
	}
}

func mapAssignmentSubmissionPtr(v *svc.AssignmentSubmission) *AssignmentSubmission {
	if v == nil {
		return nil
	}
	r := mapAssignmentSubmission(*v)
	return &r
}

func mapAssignmentSubmissionSlice(v []svc.AssignmentSubmission) []AssignmentSubmission {
	if v == nil {
		return nil
	}
	r := make([]AssignmentSubmission, len(v))
	for k := range v {
		r[k] = mapAssignmentSubmission(v[k])
	}
	return r
}

type AssignmentSubmissionAttachment struct {
	ID           string `json:"id"`
	FileName     string `json:"fileName"`
	FileSize     int64  `json:"fileSize"`
	ContentType  string `json:"contentType"`
	PresignedURL string `json:"presignedUrl"`
}

func mapAssignmentSubmissionAttachment(v svc.AssignmentSubmissionAttachment) AssignmentSubmissionAttachment {
	return AssignmentSubmissionAttachment{
		ID:           v.ID,
		FileName:     v.FileName,
		FileSize:     v.FileSize,
		ContentType:  v.ContentType,
		PresignedURL: v.PresignedURL,
	}
}

func mapAssignmentSubmissionAttachmentPtr(v *svc.AssignmentSubmissionAttachment) *AssignmentSubmissionAttachment {
	if v == nil {
		return nil
	}
	r := mapAssignmentSubmissionAttachment(*v)
	return &r
}

func mapAssignmentSubmissionAttachmentSlice(v []svc.AssignmentSubmissionAttachment) []AssignmentSubmissionAttachment {
	if v == nil {
		return nil
	}
	r := make([]AssignmentSubmissionAttachment, len(v))
	for k := range v {
		r[k] = mapAssignmentSubmissionAttachment(v[k])
	}
	return r
}
