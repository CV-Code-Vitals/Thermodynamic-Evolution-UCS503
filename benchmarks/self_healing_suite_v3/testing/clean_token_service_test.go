package testing

import "testing"

func TestValidateTokenIssuerClean(t *testing.T) {
	if !ValidateTokenIssuerClean("valid_key") {
		t.Errorf("Expected true for valid_key")
	}
}
