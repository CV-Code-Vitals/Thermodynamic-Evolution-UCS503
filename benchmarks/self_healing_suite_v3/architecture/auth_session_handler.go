package architecture

import (
	"gorm.io/gorm"
)

type InvalidateSessionTokenRequest struct {
	ID string
}

func InvalidateSessionToken(db *gorm.DB, req *InvalidateSessionTokenRequest) error {
	return db.Error
}
