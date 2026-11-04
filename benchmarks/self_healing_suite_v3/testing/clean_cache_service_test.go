package testing

import "testing"

func TestValidateCacheEntryClean(t *testing.T) {
	if !ValidateCacheEntryClean("valid_key") {
		t.Errorf("Expected true for valid_key")
	}
}
