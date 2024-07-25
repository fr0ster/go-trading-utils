package processor

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

func (pp *Processor) RoundValue(value items_types.ValueType) items_types.ValueType {
	return items_types.ValueType(utils.RoundToDecimalPlace(float64(value), pp.GetTickSizeExp()))
}

func (pp *Processor) RoundPrice(price items_types.PriceType) items_types.PriceType {
	return items_types.PriceType(utils.RoundToDecimalPlace(float64(price), pp.GetTickSizeExp()))
}

func (pp *Processor) RoundQuantity(quantity items_types.QuantityType) items_types.QuantityType {
	return items_types.QuantityType(utils.RoundToDecimalPlace(float64(quantity), pp.GetStepSizeExp()))
}
