package portal

import (
	svc "app/internal/application/portal/uploads"
	"errors"
	"log"
	"net/http"
)

func writeUploadsError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, svc.ErrForbidden):
		w.WriteHeader(http.StatusForbidden)
	case errors.Is(err, svc.ErrUnprocessableEntity):
		w.WriteHeader(http.StatusUnprocessableEntity)
	case errors.Is(err, svc.ErrBadRequest):
		w.WriteHeader(http.StatusBadRequest)
	default:
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
