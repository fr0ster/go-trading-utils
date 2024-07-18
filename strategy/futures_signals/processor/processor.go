package processor

import (
	"fmt"

	types "github.com/fr0ster/go-trading-utils/types/depth/items"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	"github.com/sirupsen/logrus"
)

func (pp *PairProcessor) CalcValueForQuantity(
	P1 types.PriceType,
	Q1 types.QuantityType,
	P2 types.PriceType) (
	value types.PriceType,
	quantity types.QuantityType,
	middlePrice types.PriceType,
	n int) {
	var (
		deltaPrice types.PriceType
	)
	if P1 < P2 {
		deltaPrice = types.PriceType(pp.GetDeltaPrice())
	} else {
		deltaPrice = types.PriceType(-pp.GetDeltaPrice())
	}
	n = pp.FindLengthOfProgression(float64(P1), float64(P1)*(1+float64(deltaPrice)), float64(P2))
	value = types.PriceType(pp.Sum(
		float64(P1)*float64(Q1),
		pp.GetDelta(
			float64(P1)*float64(Q1),
			float64(P1)*(1+float64(deltaPrice))*float64(Q1)*(1+float64(pp.GetDeltaQuantity()))),
		n))
	quantity = types.QuantityType(pp.Sum(float64(Q1), pp.GetDelta(float64(Q1), float64(Q1)*(1+float64(pp.GetDeltaQuantity()))), n))
	middlePrice = types.PriceType(pp.RoundPrice(types.PriceType(float64(value) / float64(quantity))))
	return
}

func (pp *PairProcessor) CalculateInitialPosition(
	buyPrice,
	endPrice types.PriceType) (
	value types.PriceType,
	quantity types.QuantityType,
	middlePrice types.PriceType,
	initialQuantity types.QuantityType, n int, err error) {
	low := types.QuantityType(pp.RoundQuantity(types.QuantityType(pp.notional / float64(buyPrice))))
	high := types.QuantityType(pp.RoundQuantity(types.QuantityType(float64(pp.limitOnPosition) * float64(pp.leverage) / float64(buyPrice))))

	for pp.RoundQuantity(high-low) > types.QuantityType(pp.StepSize) {
		mid := types.QuantityType(pp.RoundQuantity((low + high) / 2))
		value, _, _, n = pp.CalcValueForQuantity(buyPrice, mid, endPrice)
		if value <= types.PriceType(float64(pp.limitOnPosition)*float64(pp.leverage)) && n >= pp.minSteps {
			low = mid
		} else {
			high = mid
		}
	}

	value, quantity, middlePrice, n = pp.CalcValueForQuantity(buyPrice, high, endPrice)
	if value < types.PriceType(float64(pp.limitOnPosition)*float64(pp.leverage)) && n >= pp.minSteps {
		initialQuantity = types.QuantityType(pp.RoundQuantity(high))
		return
	}
	value, quantity, middlePrice, n = pp.CalcValueForQuantity(buyPrice, low, endPrice)
	if value < types.PriceType(float64(pp.limitOnPosition)*float64(pp.leverage)) && n >= pp.minSteps {
		initialQuantity = pp.RoundQuantity(low)
		return
	}

	err = fmt.Errorf("can't calculate initial position")
	return
}

func (pp *PairProcessor) InitPositionGridUp(price types.PriceType) (
	valueUp types.PriceType,
	quantityUp types.QuantityType,
	middlePriceUp types.PriceType,
	startQuantityUp types.QuantityType,
	stepsUp int,
	err error) {
	var (
		priceUp           types.PriceType
		currentQuantityUp types.QuantityType
	)
	valueUp, quantityUp, middlePriceUp, startQuantityUp, stepsUp, err = pp.CalculateInitialPosition(
		price,
		pp.UpBound)
	if err != nil {
		return
	}
	if float64(startQuantityUp)*float64(price) < pp.notional {
		err = fmt.Errorf("we need more money for position if price gone up: %v but can buy only for %v",
			pp.notional, float64(startQuantityUp)*float64(price))
		return
	}
	pp.up.Clear(false)
	for i := 1; i <= stepsUp; i++ {
		priceUp = pp.RoundPrice(types.PriceType(pp.FindNthTerm(float64(price), float64(price)*(1+float64(pp.GetDeltaPrice())), i+1)))
		currentQuantityUp = pp.RoundQuantity(
			types.QuantityType(
				pp.FindNthTerm(
					float64(startQuantityUp),
					float64(startQuantityUp)*(1+float64(pp.GetDeltaQuantity())), i+1)))
		if float64(currentQuantityUp)*float64(price) < pp.notional {
			err = fmt.Errorf("we need more money for position if price gone up: %v but can buy only for %v",
				pp.notional, float64(currentQuantityUp)*float64(price))
			return
		}
		pp.up.ReplaceOrInsert(&pair_price_types.PairPrice{Price: priceUp, Quantity: currentQuantityUp})
	}
	return

}

func (pp *PairProcessor) InitPositionGridDown(price types.PriceType) (
	valueDown types.PriceType,
	quantityDown types.QuantityType,
	middlePriceDown types.PriceType,
	startQuantityDown types.QuantityType,
	stepsDown int,
	err error) {
	var (
		priceDown           types.PriceType
		currentQuantityDown types.QuantityType
	)
	valueDown, quantityDown, middlePriceDown, startQuantityDown, stepsDown, err = pp.CalculateInitialPosition(
		price,
		pp.LowBound)
	if err != nil {
		return
	}
	if float64(currentQuantityDown)*float64(price) < pp.notional {
		err = fmt.Errorf("we need more money for position if price gone down: %v but can buy only for %v",
			pp.notional, float64(currentQuantityDown)*float64(price))
		return
	}
	pp.down.Clear(false)
	for i := 1; i < stepsDown; i++ {
		priceDown = types.PriceType(pp.FindNthTerm(float64(price), float64(price)*(1-float64(pp.GetDeltaPrice())), i))
		currentQuantityDown = types.QuantityType(pp.FindNthTerm(float64(startQuantityDown), float64(startQuantityDown)*(1+float64(pp.GetDeltaQuantity())), i))
		if float64(currentQuantityDown)*float64(price) < pp.notional {
			err = fmt.Errorf("we need more money for position if price gone down: %v but can buy only for %v",
				pp.notional, float64(currentQuantityDown)*float64(price))
		}
		pp.down.ReplaceOrInsert(&pair_price_types.PairPrice{Price: priceDown, Quantity: currentQuantityDown})
	}
	return

}

func (pp *PairProcessor) NextPriceUp(price types.PriceType) types.PriceType {
	return pp.RoundPrice(price * (1 + pp.GetDeltaPrice()))
}

func (pp *PairProcessor) NextPriceDown(price types.PriceType) types.PriceType {
	return pp.RoundPrice(price * (1 - pp.GetDeltaPrice()))
}

func (pp *PairProcessor) NextQuantityUp(quantity types.QuantityType) types.QuantityType {
	return pp.RoundQuantity(quantity * (1 + pp.GetDeltaQuantity()))
}

func (pp *PairProcessor) NextQuantityDown(quantity types.QuantityType) types.QuantityType {
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

func (pp *PairProcessor) NextUp(currentPrice types.PriceType, currentQuantity ...types.QuantityType) (
	price types.PriceType,
	quantity types.QuantityType,
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

func (pp *PairProcessor) nextUps(currentPrice types.PriceType) (
	price types.PriceType,
	quantity types.QuantityType,
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

func (pp *PairProcessor) NextDown(currentPrice types.PriceType, currentQuantity ...types.QuantityType) (
	price types.PriceType,
	quantity types.QuantityType,
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

func (pp *PairProcessor) nextDowns(currentPrice types.PriceType) (
	price types.PriceType,
	quantity types.QuantityType,
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

func (pp *PairProcessor) ResetUpDown(currentPrice types.PriceType) (err error) {
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

func (pp *PairProcessor) ResetUpOrInit(currentPrice types.PriceType) (err error) {
	if pp.up.Len() == 0 && pp.down.Len() == 0 {
		_, _, _, _, _, err = pp.InitPositionGridUp(currentPrice)
	} else {
		pp.ResetUpDown(currentPrice)
	}
	return
}

func (pp *PairProcessor) ResetDownOrInit(currentPrice types.PriceType) (err error) {
	if pp.up.Len() == 0 && pp.down.Len() == 0 {
		_, _, _, _, _, err = pp.InitPositionGridDown(currentPrice)
	} else {
		pp.ResetUpDown(currentPrice)
	}
	return
}
