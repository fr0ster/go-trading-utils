package info

type (
	SymbolPrice float64
	SymbolName  string
	PriceRecord struct {
		SymbolName SymbolName
		Price      SymbolPrice
	}
)
