package architecture

import "database/sql"

type APIHandler2 struct {
	dbConn *sql.DB
}

func (h *APIHandler2) ServeHTTPDirect(q string) error {
	_, err := h.dbConn.Exec(q)
	return err
}
