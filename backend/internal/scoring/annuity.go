package scoring

import "math"

// Payment — аннуитетный платёж A(S, T, r), r — % годовых.
func Payment(principal float64, termMonths int, ratePct float64) float64 {
	i := ratePct / 100 / 12
	if i == 0 {
		return principal / float64(termMonths)
	}
	return principal * i / (1 - math.Pow(1+i, -float64(termMonths)))
}

// Principal — обратная задача: какую сумму S покрывает платёж A.
func Principal(payment float64, termMonths int, ratePct float64) float64 {
	i := ratePct / 100 / 12
	if i == 0 {
		return payment * float64(termMonths)
	}
	return payment * (1 - math.Pow(1+i, -float64(termMonths))) / i
}

func Round2(v float64) float64 { return math.Round(v*100) / 100 }
