package portal

import (
	svc "app/internal/application/portal/posts"
	"errors"
	"log"
	"net/http"
)

func writePostsError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, svc.ErrForbidden):
		w.WriteHeader(http.StatusForbidden)
	default:
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
