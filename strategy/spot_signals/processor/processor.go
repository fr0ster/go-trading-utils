package processor

func (pp *PairProcessor) NextPriceUp(price float64) float64 {
	return pp.RoundPrice(price * (1 + pp.GetDeltaPrice()))
}

func (pp *PairProcessor) NextPriceDown(price float64) float64 {
	return pp.RoundPrice(price * (1 - pp.GetDeltaPrice()))
}

func (pp *PairProcessor) NextQuantityUp(quantity float64) float64 {
	return pp.RoundQuantity(quantity * (1 + pp.GetDeltaQuantity()))
}

func (pp *PairProcessor) NextQuantityDown(quantity float64) float64 {
	return pp.RoundQuantity(quantity * (1 - pp.GetDeltaQuantity()))
}
