package academicyears

import "errors"

var (
	ErrForbidden            = errors.New("forbidden")
	ErrUnprocessableEntity  = errors.New("unprocessable entity")
	ErrAcademicYearConflict = errors.New("academic year conflict")
)

var ErrYearIsActive = errors.New("year is active")

var ErrYearHasGrades = errors.New("year has grades")

var ErrNoActiveYear = errors.New("no active year")

var ErrBadRequest = errors.New("bad request")
