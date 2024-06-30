package processor

import (
	"fmt"

	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
)

func (pp *PairProcessor) CalcValueForQuantity(
	P1 float64,
	Q1 float64,
	P2 float64) (
	value float64,
	quantity float64,
	middlePrice float64,
	n int) {
	var (
		deltaPrice float64
	)
	if P1 < P2 {
		deltaPrice = pp.GetDeltaPrice()
	} else {
		deltaPrice = -pp.GetDeltaPrice()
	}
	n = pp.FindLengthOfProgression(P1, P1*(1+deltaPrice), P2)
	value = pp.Sum(P1*Q1, pp.GetDelta(P1*Q1, P1*(1+deltaPrice)*Q1*(1+pp.GetDeltaQuantity())), n)
	quantity = pp.Sum(Q1, pp.GetDelta(Q1, Q1*(1+pp.GetDeltaQuantity())), n)
	middlePrice = pp.RoundPrice(value / quantity)
	return
}

func (pp *PairProcessor) CalculateInitialPosition(
	buyPrice,
	endPrice float64) (
	value,
	quantity,
	middlePrice,
	initialQuantity float64, n int, err error) {
	low := pp.RoundQuantity(pp.notional / buyPrice)
	high := pp.RoundQuantity(pp.limitOnPosition * float64(pp.leverage) / buyPrice)

	for pp.RoundQuantity(high-low) > pp.stepSizeDelta {
		mid := pp.RoundQuantity((low + high) / 2)
		value, _, _, n = pp.CalcValueForQuantity(buyPrice, mid, endPrice)
		if value <= pp.limitOnPosition*float64(pp.leverage) && n >= pp.minSteps {
			low = mid
		} else {
			high = mid
		}
	}

	value, quantity, middlePrice, n = pp.CalcValueForQuantity(buyPrice, high, endPrice)
	if value < pp.limitOnPosition*float64(pp.leverage) && n >= pp.minSteps {
		initialQuantity = pp.RoundQuantity(high)
		return
	}
	value, quantity, middlePrice, n = pp.CalcValueForQuantity(buyPrice, low, endPrice)
	if value < pp.limitOnPosition*float64(pp.leverage) && n >= pp.minSteps {
		initialQuantity = pp.RoundQuantity(low)
		return
	}

	err = fmt.Errorf("can't calculate initial position")
	return
}

func (pp *PairProcessor) InitPositionGrid(price float64) (
	valueUp float64,
	quantityUp float64,
	middlePriceUp float64,
	startQuantityUp float64,
	stepsUp int,
	valueDown,
	quantityDown float64,
	middlePriceDown float64,
	startQuantityDown float64,
	stepsDown int,
	err error) {
	var (
		priceUp             float64
		currentQuantityUp   float64
		priceDown           float64
		currentQuantityDown float64
	)
	valueUp, quantityUp, middlePriceUp, startQuantityUp, stepsUp, err = pp.CalculateInitialPosition(
		price,
		pp.UpBound)
	if err != nil {
		return
	}
	priceUp = price * (1 + pp.GetDeltaPrice())
	currentQuantityUp = startQuantityUp
	pp.up.Clear(false)
	for i := 2; i < stepsUp; i++ {
		pp.up.ReplaceOrInsert(&pair_price_types.PairPrice{Price: priceUp, Quantity: currentQuantityUp})
		priceUp = pp.RoundPrice(pp.FindNthTerm(priceUp, priceUp*(1+pp.GetDeltaPrice()), i+1))
		currentQuantityUp = pp.RoundQuantity(pp.FindNthTerm(currentQuantityUp, currentQuantityUp*(1+pp.GetDeltaQuantity()), i+1))
	}
	valueDown, quantityDown, middlePriceDown, startQuantityDown, stepsDown, err = pp.CalculateInitialPosition(
		price,
		pp.LowBound)
	if err != nil {
		return
	}
	priceDown = price * (1 - pp.GetDeltaPrice())
	currentQuantityDown = startQuantityDown
	pp.down.Clear(false)
	for i := 2; i < stepsUp; i++ {
		pp.down.ReplaceOrInsert(&pair_price_types.PairPrice{Price: priceDown, Quantity: currentQuantityDown})
		priceDown = pp.FindNthTerm(priceDown, priceDown*(1-pp.GetDeltaPrice()), i+1)
		currentQuantityDown = pp.FindNthTerm(currentQuantityDown, currentQuantityDown*(1+pp.GetDeltaQuantity()), i+1)
	}
	if quantityUp*price < pp.notional {
		err = fmt.Errorf("we need more money for position if price gone up: %v but can buy only for %v", pp.notional, quantityUp*price)
	}
	if currentQuantityDown*price < pp.notional {
		err = fmt.Errorf("we need more money for position if price gone down: %v but can buy only for %v", pp.notional, currentQuantityDown*price)
	}
	return

}

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

func (pp *PairProcessor) GetUpLength() int {
	return pp.up.Len()
}

func (pp *PairProcessor) GetDownLength() int {
	return pp.down.Len()
}

func (pp *PairProcessor) UpDownClear() {
	pp.up.Clear(false)
	pp.down.Clear(false)
}

func (pp *PairProcessor) NextUp(currentPrice float64, currentQuantity ...float64) (price, quantity float64, err error) {
	if val := pp.up.Min(); val != nil {
		pair := val.(*pair_price_types.PairPrice)
		pp.up.Delete(val)
		if len(currentQuantity) == 1 {
			pp.down.ReplaceOrInsert(&pair_price_types.PairPrice{Price: currentPrice, Quantity: currentQuantity[0]})
		} else if len(currentQuantity) == 0 {
			pp.down.ReplaceOrInsert(&pair_price_types.PairPrice{Price: currentPrice, Quantity: pair.Quantity})
		}
		return pair.Price, pair.Quantity, nil
	} else {
		return 0, 0, fmt.Errorf("can't get next up price")
	}
}

func (pp *PairProcessor) nextUps(currentPrice float64) (price, quantity float64, err error) {
	for {
		price, quantity, err = pp.NextUp(currentPrice)
		if err != nil {
			return
		}
		if price > currentPrice {
			return
		}
	}
}

func (pp *PairProcessor) NextDown(currentPrice float64, currentQuantity ...float64) (price, quantity float64, err error) {
	if val := pp.down.Max(); val != nil {
		pair := val.(*pair_price_types.PairPrice)
		pp.down.Delete(val)
		if len(currentQuantity) == 1 {
			pp.up.ReplaceOrInsert(&pair_price_types.PairPrice{Price: currentPrice, Quantity: currentQuantity[0]})
		} else if len(currentQuantity) == 0 {
			pp.up.ReplaceOrInsert(&pair_price_types.PairPrice{Price: currentPrice, Quantity: pair.Quantity})
		}
		return pair.Price, pair.Quantity, nil
	} else {
		return 0, 0, fmt.Errorf("can't get next down price")
	}
}

func (pp *PairProcessor) nextDowns(currentPrice float64) (price, quantity float64, err error) {
	for {
		price, quantity, err = pp.NextDown(currentPrice)
		if err != nil {
			return
		}
		if price < currentPrice {
			return
		}
	}
}

func (pp *PairProcessor) ResetUpDown(currentPrice float64) (err error) {
	if pp.up.Len() == 0 && pp.down.Len() == 0 {
		_, _, _, _, _, _, _, _, _, _, err = pp.InitPositionGrid(currentPrice)
	} else {
		up := pp.up.Min()
		down := pp.down.Max()
		if up != nil && down != nil {
			upPrice := up.(*pair_price_types.PairPrice).Price
			downPrice := down.(*pair_price_types.PairPrice).Price
			if currentPrice < upPrice && currentPrice > downPrice {
				return
			} else if currentPrice >= upPrice {
				_, _, err = pp.nextUps(currentPrice)
			} else if currentPrice <= downPrice {
				_, _, err = pp.nextDowns(currentPrice)
			}
		} else if up == nil && currentPrice <= down.(*pair_price_types.PairPrice).Price {
			_, _, err = pp.nextDowns(currentPrice)
		} else if down == nil && currentPrice >= up.(*pair_price_types.PairPrice).Price {
			_, _, err = pp.nextUps(currentPrice)
		}
	}
	return
}
