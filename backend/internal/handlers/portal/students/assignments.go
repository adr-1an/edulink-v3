package students

import (
	svc "app/internal/application/portal/students"
	"app/internal/auth"
	"app/internal/portalusers"

	"encoding/json"
	"net/http"
)

func (h *Handler) ListAllAssignmentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUID(w, r, h.DB, ctx, portalusers.AccTypeStudent)
	if err != nil {
		return
	}
	res, err := h.App.Portal.Students.ListAllAssignments(ctx, &svc.ListAllAssignmentsInput{UserID: userID})
	if err != nil {
		writeStudentsError(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"assignments": mapListAllAssignmentsAssignmentSlice(res.Assignments)})
}
