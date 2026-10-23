package scoring

import (
	"math"
	"testing"
)

func TestPaymentPrincipalRoundTrip(t *testing.T) {
	a := Payment(2_000_000, 12, 19.5)
	if math.Abs(a-184_790.75) > 1 {
		t.Fatalf("payment = %.2f", a)
	}
	if s := Principal(a, 12, 19.5); math.Abs(s-2_000_000) > 0.01 {
		t.Fatalf("principal = %.2f", s)
	}
}
