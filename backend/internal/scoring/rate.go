package scoring

// EquilibriumRate — r*: ставка, при которой ожидаемый доход покрывает ожидаемые потери,
// упрощённо r* = (r₀ + PD·LGD) / (1 − PD).
// TODO(математик): заменить на формулу задачи 8.
func EquilibriumRate(baseRatePct, pd, lgd float64) float64 {
	return Round2((baseRatePct/100 + pd*lgd) / (1 - pd) * 100)
}
