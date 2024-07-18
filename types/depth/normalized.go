package depth

// func (d *Depths) GetNormalizedAsk(price types.PriceType) (item *types.NormalizedItem, err error) {
// 	if val := d.askNormalized.Get(d.NewAskNormalizedItem(price)); val != nil {
// 		item = val.(*types.NormalizedItem)
// 	}
// 	return
// }

// func (d *Depths) GetNormalizedAskSumma(price types.PriceType) (summa, summaTest types.QuantityType) {
// 	if d.askNormalized != nil {
// 		askN, _ := d.GetNormalizedAsk(price)
// 		if askN == nil {
// 			return
// 		}
// 		summaTest = askN.GetQuantity()
// 		askN.GetDepths().Ascend(func(i btree.Item) bool {
// 			summa += i.(*types.DepthItem).GetQuantity()
// 			return true
// 		})
// 	}
// 	return
// }

// func (d *Depths) GetNormalizedAsks() *btree.BTree {
// 	return d.askNormalized
// }

// func (d *Depths) GetNormalizedAsksSummaAll() (summa types.QuantityType) {
// 	if d.askNormalized != nil {
// 		d.askNormalized.Ascend(func(i btree.Item) bool {
// 			summa += i.(*types.NormalizedItem).GetQuantity()
// 			return true
// 		})
// 	}
// 	return
// }

// func (d *Depths) GetNormalizedBid(price types.PriceType) (item *types.NormalizedItem, err error) {
// 	if val := d.bidNormalized.Get(d.NewBidNormalizedItem(price)); val != nil {
// 		item = val.(*types.NormalizedItem)
// 	}
// 	return
// }

// func (d *Depths) GetNormalizedBids() *btree.BTree {
// 	return d.bidNormalized
// }

// func (d *Depths) GetNormalizedBidSumma(price types.PriceType) (summa, summaTest types.QuantityType) {
// 	if d.bidNormalized != nil {
// 		bidN, _ := d.GetNormalizedBid(price)
// 		if bidN == nil {
// 			return
// 		}
// 		summaTest = bidN.GetQuantity()
// 		bidN.GetDepths().Ascend(func(i btree.Item) bool {
// 			summa += i.(*types.DepthItem).GetQuantity()
// 			return true
// 		})
// 	}
// 	return
// }

// func (d *Depths) GetNormalizedBidsSummaAll() (summa types.QuantityType) {
// 	if d.bidNormalized != nil {
// 		d.bidNormalized.Ascend(func(i btree.Item) bool {
// 			summa += i.(*types.NormalizedItem).GetQuantity()
// 			return true
// 		})
// 	}
// 	return
// }

// func (d *Depths) addNormalized(tree *btree.BTree, price types.PriceType, quantity types.QuantityType, RoundUp bool) (err error) {
// 	if tree != nil {
// 		if old := tree.Get(d.newNormalizedItem(price, RoundUp)); old != nil {
// 			// MinMax && Depths
// 			old.(*types.NormalizedItem).Add(price, quantity)
// 		} else {
// 			item := d.newNormalizedItem(price, RoundUp, quantity)
// 			// item.Add(price, quantity)
// 			tree.ReplaceOrInsert(item)
// 		}
// 	} else {
// 		err = errors.New("tree is nil")
// 	}
// 	return
// }

// func (d *Depths) AddAskNormalized(price types.PriceType, quantity types.QuantityType) error {
// 	return d.addNormalized(d.askNormalized, price, quantity, true)
// }

// func (d *Depths) AddBidNormalized(price types.PriceType, quantity types.QuantityType) error {
// 	return d.addNormalized(d.bidNormalized, price, quantity, false)
// }

// func (d *Depths) deleteNormalized(tree *btree.BTree, price types.PriceType, quantity types.QuantityType, roundUp bool) (err error) {
// 	if tree != nil {
// 		if old := tree.Get(d.newNormalizedItem(price, roundUp)); old != nil {
// 			old.(*types.NormalizedItem).Delete(price, quantity)
// 		}
// 	} else {
// 		err = errors.New("tree is nil")
// 	}
// 	return
// }

// func (d *Depths) DeleteAskNormalized(price types.PriceType, quantity types.QuantityType) (err error) {
// 	err = d.deleteNormalized(d.askNormalized, price, quantity, true)
// 	if err != nil {
// 		return
// 	}
// 	if old := d.askNormalized.Get(d.NewAskNormalizedItem(price)); old != nil {
// 		if old.(*types.NormalizedItem).IsShouldDelete() {
// 			d.askNormalized.Delete(d.NewAskNormalizedItem(price))
// 		}
// 	}
// 	return
// }

// func (d *Depths) DeleteBidNormalized(price types.PriceType, quantity types.QuantityType) (err error) {
// 	err = d.deleteNormalized(d.bidNormalized, price, quantity, false)
// 	if err != nil {
// 		return
// 	}
// 	if old := d.askNormalized.Get(d.NewBidNormalizedItem(price)); old != nil {
// 		if old.(*types.NormalizedItem).IsShouldDelete() {
// 			d.bidNormalized.Delete(d.NewBidNormalizedItem(price))
// 		}
// 	}
// 	return
// }

// func (d *Depths) newNormalizedItem(price types.PriceType, roundUp bool, quantity ...types.QuantityType) (normalized *types.NormalizedItem) {
// 	if len(quantity) > 0 {
// 		normalized = types.NewNormalizedItem(price, d.degree, d.expBase, roundUp, quantity[0])
// 	} else {
// 		normalized = types.NewNormalizedItem(price, d.degree, d.expBase, roundUp)
// 	}
// 	return
// }

// func (d *Depths) NewAskNormalizedItem(price types.PriceType, quantity ...types.QuantityType) (normalized *types.NormalizedItem) {
// 	normalized = d.newNormalizedItem(price, true, quantity...)
// 	return
// }

// func (d *Depths) NewBidNormalizedItem(price types.PriceType, quantity ...types.QuantityType) (normalized *types.NormalizedItem) {
// 	normalized = d.newNormalizedItem(price, false, quantity...)
// 	return
// }
