package domain

// Product — продукт и его кредитная политика (решение отделено от скоринга).
type Product struct {
	ID               string
	Code             string
	Name             string
	Description      string
	MinAmount        float64
	MaxAmount        float64
	MinTermMonths    int
	MaxTermMonths    int
	BaseRate         float64 // r₀, % годовых
	LGD              float64
	PDAutoApproveMax float64 // 0.10
	PDRejectMin      float64 // 0.25
	GapThresholdQ    float64 // допустимая P_gap
	PaymentShareK    float64 // доля потока на платежи
}

func (p Product) Validate(amount float64, term int) error {
	if amount < p.MinAmount || amount > p.MaxAmount {
		return Invalid("сумма вне границ продукта")
	}
	if term < p.MinTermMonths || term > p.MaxTermMonths {
		return Invalid("срок вне границ продукта")
	}
	return nil
}
