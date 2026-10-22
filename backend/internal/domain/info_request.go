package domain

type InfoOrigin string

const (
	OriginAutoCheck InfoOrigin = "AUTO_CHECK"
	OriginInspector InfoOrigin = "INSPECTOR"
)

type InfoRequest struct {
	ID            string
	ApplicationID string
	Origin        InfoOrigin
	RequestedBy   string
	Description   string
	Status        string // OPEN | FULFILLED | EXPIRED | CANCELLED
}

// ReturnStatus — куда возвращается заявка после дозагрузки (state.mmd).
func (r InfoRequest) ReturnStatus() Status {
	if r.Origin == OriginInspector {
		return StatusManualReview
	}
	return StatusSubmitted
}
