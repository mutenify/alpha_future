package domain

// Scorecard — активная версия балльной карты (scorecard_models + scorecard_bins).
type Scorecard struct {
	Version   string
	BaseScore int
	BaseOdds  float64
	PDO       int
	ScoreMin  int
	ScoreMax  int
	Bins      []Bin
}

// Bin — числовой интервал (Lower, Upper] (nil — бесконечность) или категория.
type Bin struct {
	Factor      string
	Lower       *float64
	Upper       *float64
	Category    *string
	Points      int
	Explanation string
}

func (b Bin) Matches(v float64) bool {
	return (b.Lower == nil || v > *b.Lower) && (b.Upper == nil || v <= *b.Upper)
}

type Industry struct {
	Code      string
	Name      string
	Margin    float64
	RevenueCV float64
}

type Rules struct {
	Products   []Product
	Scorecard  Scorecard
	Industries []Industry
}
