package repository

import "testing"

func TestOrderRepoInit(t *testing.T) {
	repo := NewOrderRepository(nil)
	if repo == nil {
		t.Errorf("Expected repo not nil")
	}
}
