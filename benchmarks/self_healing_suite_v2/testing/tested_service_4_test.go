package testing
import "testing"
func TestCoveredCalculateSum4(t *testing.T) {
	if CoveredCalculateSum4(2, 3) != 5 { t.Fail() }
}
