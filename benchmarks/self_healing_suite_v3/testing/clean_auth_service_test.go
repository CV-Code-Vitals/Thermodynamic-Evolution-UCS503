package testing

import "testing"

func TestValidateAuthSessionClean(t *testing.T) {
	if !ValidateAuthSessionClean("valid_key") {
		t.Errorf("Expected true for valid_key")
	}
}
