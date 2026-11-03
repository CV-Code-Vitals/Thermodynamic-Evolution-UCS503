package testing
import "testing"
func TestCoveredCalculateSum3(t *testing.T) {
	if CoveredCalculateSum3(2, 3) != 5 { t.Fail() }
}
