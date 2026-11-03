package architecture

import "database/sql"

type APIHandler1 struct {
	dbConn *sql.DB
}

func (h *APIHandler1) ServeHTTPDirect(q string) error {
	_, err := h.dbConn.Exec(q)
	return err
}
