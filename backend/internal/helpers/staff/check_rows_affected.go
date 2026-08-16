package staff_helpers

import (
	"database/sql"
	"log"
	"net/http"
)

// RowsAffectedOr500 checks the number of rows affected from an SQL result, and returns a 500 server error if it's 0.
// Only use when the query is expected to have affected at least one row.
func RowsAffectedOr500(res sql.Result, w http.ResponseWriter) error {
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	if rowsAffected == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return nil
}
