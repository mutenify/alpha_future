package scoring

import (
	"encoding/csv"
	"math"
	"os"
	"strconv"
	"testing"

	"smartcredit/internal/domain"
)

const numericTolerance = 1e-8

func TestGolden(t *testing.T) {
	f, err := os.Open("testdata/golden_cases.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 11 {
		t.Fatalf("golden rows=%d, want 10", len(rows)-1)
	}
	header := map[string]int{}
	for i, name := range rows[0] {
		header[name] = i
	}
	for _, row := range rows[1:] {
		caseID := value(row, header, "case_id")
		t.Run("case_"+caseID, func(t *testing.T) {
			if value(row, header, "special") != "" {
				return
			}
			profile := domain.FinancialProfile{
				BusinessAgeMonths: integer(row, header, "age_months"), Overdues12m: integer(row, header, "overdues12m"),
				MaxOverdueDays: integer(row, header, "max_overdue_days"), Inquiries3m: integer(row, header, "inquiries3m"),
				TaxDebt: number(row, header, "tax_debt"), RevenueTrend: number(row, header, "revenue_trend"),
				RevenueCV: number(row, header, "revenue_cv"), IndustryCode: value(row, header, "industry_code"),
				CashFlowMu: number(row, header, "mu"), CashFlowSigma: number(row, header, "sigma"),
				CurrentDebtPayments: number(row, header, "debt_old"), CashBalance: number(row, header, "cash_balance"),
			}
			product := goldenProduct(value(row, header, "product_code"))
			amount, term := number(row, header, "amount"), integer(row, header, "term_months")
			points, pd, _ := Score(profile, goldenScorecard())
			equalInt(t, caseID, "score", points, integer(row, header, "score"))
			closeValue(t, caseID, "pd", pd, number(row, header, "pd"), numericTolerance)
			rate := EquilibriumRate(product.BaseRate, pd, product.LGD)
			closeValue(t, caseID, "rate", rate, number(row, header, "rate"), numericTolerance)
			monthly := Payment(amount, term, rate)
			closeValue(t, caseID, "payment", monthly, number(row, header, "payment"), numericTolerance)
			gap := PGap(profile.CashFlowMu, profile.CashFlowSigma, profile.CurrentDebtPayments, monthly, profile.CashBalance, term)
			closeValue(t, caseID, "p_gap", gap, number(row, header, "p_gap"), numericTolerance)
			a, hasA := PointA(profile, product, term, rate)
			b, hasB := PointB(profile, product, amount, term, rate)
			outcome := Route(pd, product, Fits(profile, product, amount, term, rate))
			if outcome == domain.OutcomeCounterOffer && !hasA && !hasB {
				outcome = domain.OutcomeManualReview
			}
			if got, want := string(outcome), value(row, header, "expected_status"); got != want {
				t.Fatalf("case=%s field=status expected=%s actual=%s", caseID, want, got)
			}
			checkTerms(t, caseID, "point_a", a, hasA, row, header)
			checkTerms(t, caseID, "point_b", b, hasB, row, header)
		})
	}
}

func checkTerms(t *testing.T, caseID, prefix string, got Terms, ok bool, row []string, header map[string]int) {
	wantAmount := value(row, header, prefix+"_amount")
	if wantAmount == "" {
		if ok {
			t.Fatalf("case=%s field=%s expected=absent actual=%+v", caseID, prefix, got)
		}
		return
	}
	if !ok {
		t.Fatalf("case=%s field=%s expected=present actual=absent", caseID, prefix)
	}
	closeValue(t, caseID, prefix+"_amount", got.Amount, number(row, header, prefix+"_amount"), .01)
	equalInt(t, caseID, prefix+"_term", got.TermMonths, integer(row, header, prefix+"_term"))
	closeValue(t, caseID, prefix+"_payment", got.MonthlyPayment, number(row, header, prefix+"_payment"), .01)
	closeValue(t, caseID, prefix+"_p_gap", got.PGap, number(row, header, prefix+"_p_gap"), numericTolerance)
}

func closeValue(t *testing.T, caseID, field string, got, want, tolerance float64) {
	t.Helper()
	if delta := math.Abs(got - want); delta > tolerance {
		t.Fatalf("case=%s field=%s expected=%.12g actual=%.12g abs_delta=%.12g tolerance=%.12g", caseID, field, want, got, delta, tolerance)
	}
}

func equalInt(t *testing.T, caseID, field string, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("case=%s field=%s expected=%d actual=%d", caseID, field, want, got)
	}
}

func value(row []string, header map[string]int, name string) string { return row[header[name]] }
func number(row []string, header map[string]int, name string) float64 {
	v, err := strconv.ParseFloat(value(row, header, name), 64)
	if err != nil {
		panic(err)
	}
	return v
}
func integer(row []string, header map[string]int, name string) int {
	v, err := strconv.Atoi(value(row, header, name))
	if err != nil {
		panic(err)
	}
	return v
}
func f(v float64) *float64 { return &v }
func s(v string) *string   { return &v }

func goldenScorecard() domain.Scorecard {
	bin := func(name string, low, high *float64, points int) domain.Bin {
		return domain.Bin{Factor: name, Lower: low, Upper: high, Points: points}
	}
	return domain.Scorecard{Version: "v1", BaseScore: 600, BaseOdds: 50, PDO: 20, ScoreMin: 457, ScoreMax: 621, Bins: []domain.Bin{
		bin("age_months", f(0), f(12), 46), bin("age_months", f(12), f(24), 51), bin("age_months", f(24), f(60), 70), bin("age_months", f(60), nil, 66),
		bin("overdues12m", f(-1), f(0), 74), bin("overdues12m", f(0), f(1), 62), bin("overdues12m", f(1), f(3), 40), bin("overdues12m", f(3), nil, 50),
		bin("max_overdue_days", f(-1), f(0), 63), bin("max_overdue_days", f(0), f(30), 62), bin("max_overdue_days", f(30), nil, 62),
		bin("inquiries3m", f(-1), f(0), 64), bin("inquiries3m", f(0), f(2), 63), bin("inquiries3m", f(2), nil, 58),
		bin("tax_debt", f(-1), f(0), 63), bin("tax_debt", f(0), nil, 62),
		bin("revenue_trend", nil, f(-.1), 61), bin("revenue_trend", f(-.1), f(.1), 74), bin("revenue_trend", f(.1), nil, 47),
		bin("revenue_cv", f(0), f(.25), 63), bin("revenue_cv", f(.25), f(.4), 64), bin("revenue_cv", f(.4), nil, 60),
		bin("debt_load", f(-1), f(0), 72), bin("debt_load", f(0), f(.3), 62), bin("debt_load", f(.3), f(.6), 58), bin("debt_load", f(.6), nil, 44),
		{Factor: "okved_group", Category: s("10"), Points: 71}, {Factor: "okved_group", Category: s("41"), Points: 38},
		{Factor: "okved_group", Category: s("47"), Points: 77}, {Factor: "okved_group", Category: s("49"), Points: 71},
		{Factor: "okved_group", Category: s("56"), Points: 56}, {Factor: "okved_group", Category: s("62"), Points: 62},
	}}
}

func goldenProduct(code string) domain.Product {
	switch code {
	case "WORKING_CAPITAL":
		return domain.Product{Code: code, MinAmount: 300000, MaxAmount: 30000000, MinTermMonths: 3, MaxTermMonths: 36, BaseRate: 19.5, LGD: .6, PDAutoApproveMax: .1, PDRejectMin: .25, GapThresholdQ: .05, PaymentShareK: .5}
	case "INVESTMENT_LOAN":
		return domain.Product{Code: code, MinAmount: 1000000, MaxAmount: 100000000, MinTermMonths: 12, MaxTermMonths: 84, BaseRate: 18, LGD: .4, PDAutoApproveMax: .1, PDRejectMin: .25, GapThresholdQ: .03, PaymentShareK: .4}
	case "OVERDRAFT":
		return domain.Product{Code: code, MinAmount: 100000, MaxAmount: 10000000, MinTermMonths: 1, MaxTermMonths: 12, BaseRate: 21, LGD: .75, PDAutoApproveMax: .1, PDRejectMin: .25, GapThresholdQ: .08, PaymentShareK: .3}
	default:
		panic("unknown product: " + code)
	}
}
