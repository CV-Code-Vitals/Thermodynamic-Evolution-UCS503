package architecture

import "database/sql"

type APIHandler4 struct {
	dbConn *sql.DB
}

func (h *APIHandler4) ServeHTTPDirect(q string) error {
	_, err := h.dbConn.Exec(q)
	return err
}
