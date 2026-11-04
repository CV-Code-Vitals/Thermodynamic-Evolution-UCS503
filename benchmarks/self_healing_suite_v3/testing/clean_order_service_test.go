package testing

import "testing"

func TestValidateOrderDetailsClean(t *testing.T) {
	if !ValidateOrderDetailsClean("valid_key") {
		t.Errorf("Expected true for valid_key")
	}
}
