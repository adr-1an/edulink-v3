package students

import (
	svc "app/internal/application/portal/students"
	"errors"
	"log"
	"net/http"
)

func writeStudentsError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, svc.ErrForbidden):
		w.WriteHeader(http.StatusForbidden)
	case errors.Is(err, svc.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
	case errors.Is(err, svc.ErrConflict):
		w.WriteHeader(http.StatusConflict)
	case errors.Is(err, svc.ErrUnprocessableEntity):
		w.WriteHeader(http.StatusUnprocessableEntity)
	case errors.Is(err, svc.ErrBadRequest):
		w.WriteHeader(http.StatusBadRequest)
	default:
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
