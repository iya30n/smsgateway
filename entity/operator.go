package entity

type Operator struct {
	ID                 uint
	Name               string
	TotalTPS           uint
	ExpressReservedTPS uint
	IsActive           bool
}

// NormalTPS is the share of the operator's capacity left for the normal lane
// once the express lane's reservation has been carved out.
func (o Operator) NormalTPS() uint {
	if o.ExpressReservedTPS >= o.TotalTPS {
		return 0
	}

	return o.TotalTPS - o.ExpressReservedTPS
}
