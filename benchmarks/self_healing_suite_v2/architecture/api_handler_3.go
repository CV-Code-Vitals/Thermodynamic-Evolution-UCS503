package architecture

import "database/sql"

type APIHandler3 struct {
	dbConn *sql.DB
}

func (h *APIHandler3) ServeHTTPDirect(q string) error {
	_, err := h.dbConn.Exec(q)
	return err
}
