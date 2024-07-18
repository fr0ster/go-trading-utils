package depth

// // GetAsks implements depth_interface.Depths.
// func (d *Depths) GetAsks() *btree.BTree {
// 	return d.asks
// }

// // GetBids implements depth_interface.Depths.
// func (d *Depths) GetBids() *btree.BTree {
// 	return d.bids
// }

// // SetAsks implements depth_interface.Depths.
// func (d *Depths) SetAsks(asks *btree.BTree) {
// 	d.asks = asks
// 	asks.Ascend(func(i btree.Item) bool {
// 		d.asksSummaQuantity += i.(*types.DepthItem).GetQuantity()
// 		d.asksCountQuantity++
// 		d.AddAskMinMax(i.(*types.DepthItem).GetPrice(), i.(*types.DepthItem).GetQuantity())
// 		d.AddAskNormalized(i.(*types.DepthItem).GetPrice(), i.(*types.DepthItem).GetQuantity())
// 		return true
// 	})
// }

// // SetBids implements depth_interface.Depths.
// func (d *Depths) SetBids(bids *btree.BTree) {
// 	d.bids = bids
// 	bids.Ascend(func(i btree.Item) bool {
// 		d.bidsSummaQuantity += i.(*types.DepthItem).GetQuantity()
// 		d.bidsCountQuantity++
// 		d.AddBidMinMax(i.(*types.DepthItem).GetPrice(), i.(*types.DepthItem).GetQuantity())
// 		d.AddBidNormalized(i.(*types.DepthItem).GetPrice(), i.(*types.DepthItem).GetQuantity())
// 		return true
// 	})
// }

// // ClearAsks implements depth_interface.Depths.
// func (d *Depths) ClearAsks() {
// 	d.asks.Clear(false)
// 	d.asksMinMax.Clear(false)
// 	d.askNormalized.Clear(false)
// }

// // ClearBids implements depth_interface.Depths.
// func (d *Depths) ClearBids() {
// 	d.bids.Clear(false)
// 	d.bidsMinMax.Clear(false)
// 	d.bidNormalized.Clear(false)
// }

// // AskAscend implements depth_interface.Depths.
// func (d *Depths) AskAscend(iter func(btree.Item) bool) {
// 	d.asks.Ascend(iter)
// }

// // AskDescend implements depth_interface.Depths.
// func (d *Depths) AskDescend(iter func(btree.Item) bool) {
// 	d.asks.Descend(iter)
// }

// // BidAscend implements depth_interface.Depths.
// func (d *Depths) BidAscend(iter func(btree.Item) bool) {
// 	d.bids.Ascend(iter)
// }

// // BidDescend implements depth_interface.Depths.
// func (d *Depths) BidDescend(iter func(btree.Item) bool) {
// 	d.bids.Descend(iter)
// }

// func (d *Depths) GetAsksMiddleQuantity() types.QuantityType {
// 	return d.asksSummaQuantity / types.QuantityType(d.asksCountQuantity)
// }

// func (d *Depths) GetBidsMiddleQuantity() types.QuantityType {
// 	return d.bidsSummaQuantity / types.QuantityType(d.bidsCountQuantity)
// }

// func (d *Depths) GetAsksStandardDeviation() float64 {
// 	summaSquares := 0.0
// 	d.AskAscend(func(i btree.Item) bool {
// 		depth := i.(*types.DepthItem)
// 		summaSquares += depth.GetQuantityDeviation(d.GetAsksMiddleQuantity()) * depth.GetQuantityDeviation(d.GetAsksMiddleQuantity())
// 		return true
// 	})
// 	return math.Sqrt(summaSquares / float64(d.AskCount()))
// }

// func (d *Depths) GetBidsStandardDeviation() float64 {
// 	summaSquares := 0.0
// 	d.BidDescend(func(i btree.Item) bool {
// 		depth := i.(*types.DepthItem)
// 		summaSquares += depth.GetQuantityDeviation(d.GetBidsMiddleQuantity()) * depth.GetQuantityDeviation(d.GetBidsMiddleQuantity())
// 		return true
// 	})
// 	return math.Sqrt(summaSquares / float64(d.BidCount()))
// }
