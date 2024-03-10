package info

type (
	Price       float64
	DepthRecord struct {
		Price           Price
		AskLastUpdateID int64
		AskQuantity     Price
		BidLastUpdateID int64
		BidQuantity     Price
	}
)
