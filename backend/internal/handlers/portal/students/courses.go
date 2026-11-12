package students

import (
	svc "app/internal/application/portal/students"
	"app/internal/auth"
	"app/internal/portalusers"

	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListCourses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUID(w, r, h.DB, ctx, portalusers.AccTypeStudent)
	if err != nil {
		return
	}
	res, err := h.App.Portal.Students.ListCourses(ctx, &svc.ListCoursesInput{UserID: userID})
	if err != nil {
		writeStudentsError(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"courses": mapListCoursesCourseSlice(res.Courses)})
}
func (h *Handler) ViewCourseDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUID(w, r, h.DB, ctx, portalusers.AccTypeStudent)
	if err != nil {
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res, err := h.App.Portal.Students.CourseDashboard(ctx, &svc.CourseDashboardInput{UserID: userID, CourseID: courseID})
	if err != nil {
		writeStudentsError(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"course": mapCourseDashboardCourse(res.Course), "assignments": mapCourseDashboardAssignmentSlice(res.Assignments), "posts": mapCourseDashboardPostSlice(res.Posts)})
}
