package scoring

import "math"

// Phi — функция распределения стандартного нормального закона.
func Phi(x float64) float64 { return 0.5 * math.Erfc(-x/math.Sqrt2) }

// PhiInv — квантиль стандартного нормального закона.
func PhiInv(p float64) float64 { return -math.Sqrt2 * math.Erfcinv(2*p) }
