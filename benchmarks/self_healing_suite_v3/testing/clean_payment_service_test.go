package testing

import "testing"

func TestValidatePaymentAuthClean(t *testing.T) {
	if !ValidatePaymentAuthClean("valid_key") {
		t.Errorf("Expected true for valid_key")
	}
}
