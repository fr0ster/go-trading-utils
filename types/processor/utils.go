package processor

import (
	"math"

	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
)

func (pp *Processor) CeilValue(value items_types.ValueType) items_types.ValueType {
	coefficient := math.Pow10(pp.GetTickSizeExp())
	step := math.Ceil(float64(value) * coefficient)

	return items_types.ValueType(float64(step) / coefficient)
}

func (pp *Processor) FloorValue(value items_types.ValueType) items_types.ValueType {
	coefficient := math.Pow10(pp.GetTickSizeExp())
	step := math.Floor(float64(value) * coefficient)

	return items_types.ValueType(float64(step) / coefficient)
}

func (pp *Processor) RoundValue(value items_types.ValueType) items_types.ValueType {
	coefficient := math.Pow10(pp.GetTickSizeExp())
	step := math.Round(float64(value) * coefficient)

	return items_types.ValueType(float64(step) / coefficient)
}

func (pp *Processor) CeilPrice(price items_types.PriceType) items_types.PriceType {
	coefficient := math.Pow10(pp.GetTickSizeExp())
	step := math.Ceil(float64(price) * coefficient)

	return items_types.PriceType(float64(step) / coefficient)
}

func (pp *Processor) FloorPrice(price items_types.PriceType) items_types.PriceType {
	coefficient := math.Pow10(pp.GetTickSizeExp())
	step := math.Floor(float64(price) * coefficient)

	return items_types.PriceType(float64(step) / coefficient)
}

func (pp *Processor) RoundPrice(price items_types.PriceType) items_types.PriceType {
	coefficient := math.Pow10(pp.GetTickSizeExp())
	step := math.Round(float64(price) * coefficient)

	return items_types.PriceType(float64(step) / coefficient)
}

func (pp *Processor) CeilQuantity(quantity items_types.QuantityType) items_types.QuantityType {
	coefficient := math.Pow10(pp.GetStepSizeExp())
	step := math.Ceil(float64(quantity) * coefficient)

	return items_types.QuantityType(float64(step) / coefficient)
}

func (pp *Processor) FloorQuantity(quantity items_types.QuantityType) items_types.QuantityType {
	coefficient := math.Pow10(pp.GetStepSizeExp())
	step := math.Floor(float64(quantity) * coefficient)

	return items_types.QuantityType(float64(step) / coefficient)
}

func (pp *Processor) RoundQuantity(quantity items_types.QuantityType) items_types.QuantityType {
	coefficient := math.Pow10(pp.GetStepSizeExp())
	step := math.Round(float64(quantity) * coefficient)

	return items_types.QuantityType(float64(step) / coefficient)
}
