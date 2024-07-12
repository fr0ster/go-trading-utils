package depth

import (
	"errors"
)

func (d *Depth) AskMin() (min *QuantityItem, err error) {
	if d.asksMinMax.Len() == 0 {
		err = errors.New("asksMinMax is empty")
	}
	min = d.asksMinMax.Min().(*QuantityItem)
	return
}

func (d *Depth) AskMax() (max *QuantityItem, err error) {
	if d.asksMinMax.Len() == 0 {
		err = errors.New("asksMinMax is empty")
	}
	max = d.asksMinMax.Max().(*QuantityItem)
	return
}

func (d *Depth) BidMin() (min *QuantityItem, err error) {
	if d.bidsMinMax.Len() == 0 {
		err = errors.New("asksMinMax is empty")
	}
	min = d.bidsMinMax.Min().(*QuantityItem)
	return
}

func (d *Depth) BidMax() (max *QuantityItem, err error) {
	if d.bidsMinMax.Len() == 0 {
		err = errors.New("asksMinMax is empty")
	}
	max = d.bidsMinMax.Max().(*QuantityItem)
	return
}
