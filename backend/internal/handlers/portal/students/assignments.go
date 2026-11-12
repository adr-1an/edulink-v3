package students

import (
	svc "app/internal/application/portal/students"
	"app/internal/helpers"
	portal "app/internal/helpers/portal"
	"encoding/json"
	"net/http"
)

func (h *Handler) ListAllAssignmentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := portal.TokenToUID(w, r, h.DB, ctx, helpers.AccTypeStudent)
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
