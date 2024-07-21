package processor

import (
	"fmt"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	"github.com/sirupsen/logrus"
)

func (pp *PairProcessor) CalcValueForQuantity(
	P1 items_types.PriceType,
	Q1 items_types.QuantityType,
	P2 items_types.PriceType) (
	value items_types.PriceType,
	quantity items_types.QuantityType,
	middlePrice items_types.PriceType,
	n int) {
	var (
		deltaPrice items_types.PriceType
	)
	if P1 < P2 {
		deltaPrice = items_types.PriceType(pp.GetDeltaPrice())
	} else {
		deltaPrice = items_types.PriceType(-pp.GetDeltaPrice())
	}
	n = pp.FindLengthOfProgression(float64(P1), float64(P1)*(1+float64(deltaPrice)), float64(P2))
	value = items_types.PriceType(pp.Sum(
		float64(P1)*float64(Q1),
		pp.GetDelta(
			float64(P1)*float64(Q1),
			float64(P1)*(1+float64(deltaPrice))*float64(Q1)*(1+float64(pp.GetDeltaQuantity()))),
		n))
	quantity = items_types.QuantityType(pp.Sum(float64(Q1), pp.GetDelta(float64(Q1), float64(Q1)*(1+float64(pp.GetDeltaQuantity()))), n))
	middlePrice = items_types.PriceType(pp.RoundPrice(items_types.PriceType(float64(value) / float64(quantity))))
	return
}

func (pp *PairProcessor) CalculateInitialPosition(
	buyPrice,
	endPrice items_types.PriceType) (
	value items_types.PriceType,
	quantity items_types.QuantityType,
	middlePrice items_types.PriceType,
	initialQuantity items_types.QuantityType, n int, err error) {
	low := items_types.QuantityType(pp.RoundValue(pp.notional / items_types.ValueType(buyPrice)))
	high := items_types.QuantityType(pp.RoundValue(pp.limitOnPosition * items_types.ValueType(pp.leverage) / items_types.ValueType(buyPrice)))

	for pp.RoundQuantity(high-low) > items_types.QuantityType(pp.StepSize) {
		mid := items_types.QuantityType(pp.RoundQuantity((low + high) / 2))
		value, _, _, n = pp.CalcValueForQuantity(buyPrice, mid, endPrice)
		if value <= items_types.PriceType(float64(pp.limitOnPosition)*float64(pp.leverage)) && n >= pp.minSteps {
			low = mid
		} else {
			high = mid
		}
	}

	value, quantity, middlePrice, n = pp.CalcValueForQuantity(buyPrice, high, endPrice)
	if value < items_types.PriceType(float64(pp.limitOnPosition)*float64(pp.leverage)) && n >= pp.minSteps {
		initialQuantity = items_types.QuantityType(pp.RoundQuantity(high))
		return
	}
	value, quantity, middlePrice, n = pp.CalcValueForQuantity(buyPrice, low, endPrice)
	if value < items_types.PriceType(float64(pp.limitOnPosition)*float64(pp.leverage)) && n >= pp.minSteps {
		initialQuantity = pp.RoundQuantity(low)
		return
	}

	err = fmt.Errorf("can't calculate initial position")
	return
}

func (pp *PairProcessor) InitPositionGridUp(price items_types.PriceType) (
	valueUp items_types.PriceType,
	quantityUp items_types.QuantityType,
	middlePriceUp items_types.PriceType,
	startQuantityUp items_types.QuantityType,
	stepsUp int,
	err error) {
	var (
		priceUp           items_types.PriceType
		currentQuantityUp items_types.QuantityType
	)
	valueUp, quantityUp, middlePriceUp, startQuantityUp, stepsUp, err = pp.CalculateInitialPosition(
		price,
		pp.UpBound)
	if err != nil {
		return
	}
	if items_types.ValueType(float64(startQuantityUp)*float64(price)) < pp.notional {
		err = fmt.Errorf("we need more money for position if price gone up: %v but can buy only for %v",
			pp.notional, float64(startQuantityUp)*float64(price))
		return
	}
	pp.up.Clear(false)
	for i := 1; i <= stepsUp; i++ {
		priceUp = pp.RoundPrice(items_types.PriceType(pp.FindNthTerm(float64(price), float64(price)*(1+float64(pp.GetDeltaPrice())), i+1)))
		currentQuantityUp = pp.RoundQuantity(
			items_types.QuantityType(
				pp.FindNthTerm(
					float64(startQuantityUp),
					float64(startQuantityUp)*(1+float64(pp.GetDeltaQuantity())), i+1)))
		if items_types.ValueType(float64(currentQuantityUp)*float64(price)) < pp.notional {
			err = fmt.Errorf("we need more money for position if price gone up: %v but can buy only for %v",
				pp.notional, float64(currentQuantityUp)*float64(price))
			return
		}
		pp.up.ReplaceOrInsert(&pair_price_types.PairPrice{Price: priceUp, Quantity: currentQuantityUp})
	}
	return

}

func (pp *PairProcessor) InitPositionGridDown(price items_types.PriceType) (
	valueDown items_types.PriceType,
	quantityDown items_types.QuantityType,
	middlePriceDown items_types.PriceType,
	startQuantityDown items_types.QuantityType,
	stepsDown int,
	err error) {
	var (
		priceDown           items_types.PriceType
		currentQuantityDown items_types.QuantityType
	)
	valueDown, quantityDown, middlePriceDown, startQuantityDown, stepsDown, err = pp.CalculateInitialPosition(
		price,
		pp.LowBound)
	if err != nil {
		return
	}
	if items_types.ValueType(float64(currentQuantityDown)*float64(price)) < pp.notional {
		err = fmt.Errorf("we need more money for position if price gone down: %v but can buy only for %v",
			pp.notional, float64(currentQuantityDown)*float64(price))
		return
	}
	pp.down.Clear(false)
	for i := 1; i < stepsDown; i++ {
		priceDown = items_types.PriceType(pp.FindNthTerm(float64(price), float64(price)*(1-float64(pp.GetDeltaPrice())), i))
		currentQuantityDown = items_types.QuantityType(pp.FindNthTerm(float64(startQuantityDown), float64(startQuantityDown)*(1+float64(pp.GetDeltaQuantity())), i))
		if items_types.ValueType(float64(currentQuantityDown)*float64(price)) < pp.notional {
			err = fmt.Errorf("we need more money for position if price gone down: %v but can buy only for %v",
				pp.notional, float64(currentQuantityDown)*float64(price))
		}
		pp.down.ReplaceOrInsert(&pair_price_types.PairPrice{Price: priceDown, Quantity: currentQuantityDown})
	}
	return

}

func (pp *PairProcessor) GetDepth() *depth_types.Depths {
	return pp.depth
}

func (pp *PairProcessor) SetDepth(depth *depth_types.Depths) {
	pp.depth = depth
}

func (pp *PairProcessor) NextPriceUp(prices ...items_types.PriceType) items_types.PriceType {
	var err error
	if pp.depth != nil {
		return pp.RoundPrice(pp.GetDepth().NextPriceUp(items_types.PricePercentType(pp.GetDeltaPrice()), prices...))
	} else {
		price := items_types.PriceType(0.0)
		if len(prices) == 0 {
			price, err = pp.GetCurrentPrice()
			if err != nil {
				return 0
			}
		} else {
			price = prices[0]
		}
		return pp.RoundPrice(price * (1 + pp.GetDeltaPrice()))
	}
}

func (pp *PairProcessor) NextPriceDown(prices ...items_types.PriceType) items_types.PriceType {
	var err error
	if pp.depth != nil {
		return pp.RoundPrice(pp.GetDepth().NextPriceDown(items_types.PricePercentType(pp.GetDeltaPrice()), prices...))
	} else {
		price := items_types.PriceType(0.0)
		if len(prices) == 0 {
			price, err = pp.GetCurrentPrice()
			if err != nil {
				return 0
			}
		} else {
			price = prices[0]
		}
		return pp.RoundPrice(price * (1 - pp.GetDeltaPrice()))
	}
}

func (pp *PairProcessor) NextQuantityUp(quantity items_types.QuantityType) items_types.QuantityType {
	return pp.RoundQuantity(quantity * (1 + pp.GetDeltaQuantity()))
}

func (pp *PairProcessor) NextQuantityDown(quantity items_types.QuantityType) items_types.QuantityType {
	return pp.RoundQuantity(quantity * (1 - pp.GetDeltaQuantity()))
}

func (pp *PairProcessor) GetUpLength() int {
	return pp.up.Len()
}

func (pp *PairProcessor) GetDownLength() int {
	return pp.down.Len()
}

func (pp *PairProcessor) GetUpMin() *pair_price_types.PairPrice {
	return pp.up.Min().(*pair_price_types.PairPrice)
}

func (pp *PairProcessor) GetDownMax() *pair_price_types.PairPrice {
	return pp.down.Max().(*pair_price_types.PairPrice)
}

func (pp *PairProcessor) UpDownClear() {
	pp.up.Clear(false)
	pp.down.Clear(false)
}

func (pp *PairProcessor) UpDownDebug() {

	if pp.GetUpLength() != 0 {
		logrus.Debugf("Futures %s: UpLength %v, Min record: Price %v, Quantity %v",
			pp.GetPair(),
			pp.GetUpLength(),
			pp.GetUpMin().Price,
			pp.GetUpMin().Quantity)
	} else {
		logrus.Debugf("Futures %s: UpLength %v", pp.GetPair(), pp.GetUpLength())
	}
	if pp.GetDownLength() != 0 {
		logrus.Debugf("Futures %s: DownLength %v, Min record: Price %v, Quantity %v",
			pp.GetPair(),
			pp.GetDownLength(),
			pp.GetDownMax().Price,
			pp.GetDownMax().Quantity)
	} else {
		logrus.Debugf("Futures %s: DownLength %v", pp.GetPair(), pp.GetDownLength())
	}
}

func (pp *PairProcessor) NextUp(currentPrice items_types.PriceType, currentQuantity ...items_types.QuantityType) (
	price items_types.PriceType,
	quantity items_types.QuantityType,
	err error) {
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

func (pp *PairProcessor) nextUps(currentPrice items_types.PriceType) (
	price items_types.PriceType,
	quantity items_types.QuantityType,
	err error) {
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

func (pp *PairProcessor) NextDown(currentPrice items_types.PriceType, currentQuantity ...items_types.QuantityType) (
	price items_types.PriceType,
	quantity items_types.QuantityType,
	err error) {
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

func (pp *PairProcessor) nextDowns(currentPrice items_types.PriceType) (
	price items_types.PriceType,
	quantity items_types.QuantityType,
	err error) {
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

func (pp *PairProcessor) ResetUpDown(currentPrice items_types.PriceType) (err error) {
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
	return
}

func (pp *PairProcessor) ResetUpOrInit(currentPrice items_types.PriceType) (err error) {
	if pp.up.Len() == 0 && pp.down.Len() == 0 {
		_, _, _, _, _, err = pp.InitPositionGridUp(currentPrice)
	} else {
		pp.ResetUpDown(currentPrice)
	}
	return
}

func (pp *PairProcessor) ResetDownOrInit(currentPrice items_types.PriceType) (err error) {
	if pp.up.Len() == 0 && pp.down.Len() == 0 {
		_, _, _, _, _, err = pp.InitPositionGridDown(currentPrice)
	} else {
		pp.ResetUpDown(currentPrice)
	}
	return
}
