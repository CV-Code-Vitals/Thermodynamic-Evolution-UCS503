package boundary

import (
	"database/sql"
)

// DirectDatabaseBypass demonstrates presentation/API layer directly issuing raw queries,
// bypassing the service, repository, and domain transaction layers.
// Ground-truth defect: FIX-ARC-01 (Architecture: Layer boundary violation)
type DirectDatabaseBypass struct {
	rawConn *sql.DB
}

func (h *DirectDatabaseBypass) HandleUserRequest(rawQuery string) error {
	_, err := h.rawConn.Exec(rawQuery)
	return err
}

