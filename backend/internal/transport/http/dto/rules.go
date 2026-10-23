package dto

import "smartcredit/internal/domain"

type Rules struct {
	Products   []ProductPolicy `json:"products"`
	Scorecard  Scorecard       `json:"scorecard"`
	Industries []Industry      `json:"industries"`
}

type ProductPolicy struct {
	Product
	LGD              float64 `json:"lgd"`
	PDAutoApproveMax float64 `json:"pdAutoApproveMax"`
	PDRejectMin      float64 `json:"pdRejectMin"`
	GapThresholdQ    float64 `json:"gapThresholdQ"`
	PaymentShareK    float64 `json:"paymentShareK"`
}

type Scorecard struct {
	Version   string  `json:"version"`
	BaseScore int     `json:"baseScore"`
	BaseOdds  float64 `json:"baseOdds"`
	PDO       int     `json:"pdo"`
	ScoreMin  int     `json:"scoreMin"`
	ScoreMax  int     `json:"scoreMax"`
	Bins      []Bin   `json:"bins"`
}

type Bin struct {
	Factor      string   `json:"factor"`
	LowerBound  *float64 `json:"lowerBound"`
	UpperBound  *float64 `json:"upperBound"`
	Category    *string  `json:"category"`
	Points      int      `json:"points"`
	Explanation string   `json:"explanation,omitempty"`
}

type Industry struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Margin    float64 `json:"margin"`
	RevenueCV float64 `json:"revenueCv"`
}

func FromRules(r domain.Rules) Rules {
	m := r.Scorecard
	return Rules{
		Products: mapAll(r.Products, func(p domain.Product) ProductPolicy {
			return ProductPolicy{FromProduct(p), p.LGD, p.PDAutoApproveMax, p.PDRejectMin, p.GapThresholdQ, p.PaymentShareK}
		}),
		Scorecard: Scorecard{m.Version, m.BaseScore, m.BaseOdds, m.PDO, m.ScoreMin, m.ScoreMax,
			mapAll(m.Bins, func(b domain.Bin) Bin { return Bin{b.Factor, b.Lower, b.Upper, b.Category, b.Points, b.Explanation} })},
		Industries: mapAll(r.Industries, func(i domain.Industry) Industry { return Industry(i) }),
	}
}
