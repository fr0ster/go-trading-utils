package processor

func (pp *PairProcessor) GetDeltaPrice() float64 {
	return pp.deltaPrice
}

func (pp *PairProcessor) SetDeltaPrice(deltaPrice float64) {
	pp.deltaPrice = deltaPrice
}

func (pp *PairProcessor) GetDeltaQuantity() float64 {
	return pp.deltaQuantity
}

func (pp *PairProcessor) GetLimitOnTransaction() (limit float64) {
	return pp.limitOnTransaction * pp.GetFreeBalance()
}

func (pp *PairProcessor) SetBounds(price float64) {
	pp.UpBound = price * (1 + pp.UpBoundPercent)
	pp.LowBound = price * (1 - pp.LowBoundPercent)
}

func (pp *PairProcessor) GetUpBound() float64 {
	return pp.UpBound
}

func (pp *PairProcessor) GetLowBound() float64 {
	return pp.LowBound
}
