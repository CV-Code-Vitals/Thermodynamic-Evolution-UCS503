package testing
import "testing"
func TestCoveredCalculateSum2(t *testing.T) {
	if CoveredCalculateSum2(2, 3) != 5 { t.Fail() }
}
