package testing

import "testing"

func TestValidateUserProfileClean(t *testing.T) {
	if !ValidateUserProfileClean("valid_key") {
		t.Errorf("Expected true for valid_key")
	}
}
