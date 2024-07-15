package depth

import (
	"math"

	"github.com/fr0ster/go-trading-utils/utils"
)

func (d *Depth) getCoefficient(price float64) float64 {
	len := int(math.Log10(price)) + 1
	if len < d.expBase {
		return math.Pow(10, float64(d.expBase-len))
	} else {
		return 1 / math.Pow(10, float64(len-d.expBase))
	}
}

func (d *Depth) GetNormalizedPrice(price float64) float64 {
	if max := d.asks.Max().(*DepthItem); max != nil {
		coefficient := d.getCoefficient(max.Price)
		return utils.RoundToDecimalPlace(price*coefficient, d.tickSize)
	} else {
		return price
	}
}
