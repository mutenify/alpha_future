package scoring

import (
	"math"
	"testing"

	"smartcredit/internal/domain"
)

func TestPGapIncludesCashCushion(t *testing.T) {
	without := PGap(100, 20, 30, 80, 0, 12)
	with := PGap(100, 20, 30, 80, 120, 12)
	if !(with < without) {
		t.Fatalf("cash cushion must reduce PGap: without=%g with=%g", without, with)
	}
	want := Phi(0)
	if math.Abs(with-want) > 1e-12 {
		t.Fatalf("PGap=%g, want %g", with, want)
	}
}

func TestMaxPaymentIncludesCashCushionButKeepsShareLimit(t *testing.T) {
	p := domain.FinancialProfile{CashFlowMu: 1_000, CashFlowSigma: 100, CashBalance: 1_200, CurrentDebtPayments: 100}
	pr := domain.Product{GapThresholdQ: 0.05, PaymentShareK: 0.95}
	got := MaxPayment(p, pr, 12)
	want := 1_000 + 100 - 100 + 100*PhiInv(0.05)
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("MaxPayment=%g, want %g", got, want)
	}
}
