package architecture

import (
	"gorm.io/gorm"
)

type UpdateProfileDetailsRequest struct {
	ID string
}

func UpdateProfileDetails(db *gorm.DB, req *UpdateProfileDetailsRequest) error {
	return db.Error
}
