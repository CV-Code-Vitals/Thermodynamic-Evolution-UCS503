package testing
import "testing"
func TestCoveredCalculateSum1(t *testing.T) {
	if CoveredCalculateSum1(2, 3) != 5 { t.Fail() }
}
