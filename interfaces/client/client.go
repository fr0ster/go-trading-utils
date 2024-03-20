package client

import "github.com/adshao/go-binance/v2/common"

type (
	// Ask is a type alias for PriceLevel.
	Ask = common.PriceLevel

	// Bid is a type alias for PriceLevel.
	Bid = common.PriceLevel

	// DepthResponse is a type alias for common.DepthResponse.
	DepthResponse struct {
		LastUpdateID int64 `json:"lastUpdateId"`
		Time         int64 `json:"E"`
		TradeTime    int64 `json:"T"`
		Bids         []Bid `json:"bids"`
		Asks         []Ask `json:"asks"`
	}

	// CommissionRates is a type alias for common.CommissionRates.
	CommissionRates struct {
		Maker  string `json:"maker"`
		Taker  string `json:"taker"`
		Buyer  string `json:"buyer"`
		Seller string `json:"seller"`
	}

	// Balance define user balance of your account
	Balance struct {
		Asset  string `json:"asset"`
		Free   string `json:"free"`
		Locked string `json:"locked"`
	}

	// Account is a type alias for common.Account.
	Account struct {
		MakerCommission  int64           `json:"makerCommission"`
		TakerCommission  int64           `json:"takerCommission"`
		BuyerCommission  int64           `json:"buyerCommission"`
		SellerCommission int64           `json:"sellerCommission"`
		CommissionRates  CommissionRates `json:"commissionRates"`
		CanTrade         bool            `json:"canTrade"`
		CanWithdraw      bool            `json:"canWithdraw"`
		CanDeposit       bool            `json:"canDeposit"`
		UpdateTime       uint64          `json:"updateTime"`
		AccountType      string          `json:"accountType"`
		Balances         []Balance       `json:"balances"`
		Permissions      []string        `json:"permissions"`
	}
)

type Client interface {
	// NewPingService init ping service
	NewPingServiceDo() (err error)

	// NewServerTimeService init server time service
	NewServerTimeServiceDo()

	// // NewSetServerTimeService init set server time service
	// NewSetServerTimeServiceDo()

	// NewDepthService init depth service
	NewDepthServiceDo() (*DepthResponse, error)

	// // NewAggTradesService init aggregate trades service
	// NewAggTradesServiceDo()

	// // NewRecentTradesService init recent trades service
	// NewRecentTradesServiceDo()

	// // NewKlinesService init klines service
	// NewKlinesServiceDo()

	// // NewListPriceChangeStatsService init list prices change stats service
	// NewListPriceChangeStatsServiceDo()

	// // NewListPricesService init listing prices service
	// NewListPricesServiceDo()

	// // NewListBookTickersService init listing booking tickers service
	// NewListBookTickersServiceDo()

	// // NewListSymbolTickerService init listing symbols tickers
	// NewListSymbolTickerServiceDo()

	// // NewCreateOrderService init creating order service
	// NewCreateOrderServiceDo()

	// // NewCreateOCOService init creating OCO service
	// NewCreateOCOServiceDo()

	// // NewCancelOCOService init cancel OCO service
	// NewCancelOCOServiceDo()

	// // NewGetOrderService init get order service
	// NewGetOrderServiceDo()

	// // NewCancelOrderService init cancel order service
	// NewCancelOrderServiceDo()

	// // NewCancelOpenOrdersService init cancel open orders service
	// NewCancelOpenOrdersServiceDo()

	// // NewListOpenOrdersService init list open orders service
	// NewListOpenOrdersServiceDo()

	// // NewListOpenOcoService init list open oco service
	// NewListOpenOcoServiceDo()

	// // NewListOrdersService init listing orders service
	// NewListOrdersServiceDo()

	// NewGetAccountService init getting account service
	NewGetAccountServiceDo() (res *Account, err error)

	// // NewGetAPIKeyPermission init getting API key permission
	// NewGetAPIKeyPermissionDo()

	// // NewSavingFlexibleProductPositionsService get flexible products positions (Savings)
	// NewSavingFlexibleProductPositionsServiceDo()

	// // NewSavingFixedProjectPositionsService get fixed project positions (Savings)
	// NewSavingFixedProjectPositionsServiceDo()

	// // NewListSavingsFlexibleProductsService get flexible products list (Savings)
	// NewListSavingsFlexibleProductsServiceDo()

	// // NewPurchaseSavingsFlexibleProductService purchase a flexible product (Savings)
	// NewPurchaseSavingsFlexibleProductServiceDo()

	// // NewRedeemSavingsFlexibleProductService redeem a flexible product (Savings)
	// NewRedeemSavingsFlexibleProductServiceDo()

	// // NewListSavingsFixedAndActivityProductsService get fixed and activity product list (Savings)
	// NewListSavingsFixedAndActivityProductsServiceDo()

	// // NewGetAccountSnapshotService init getting account snapshot service
	// NewGetAccountSnapshotServiceDo()

	// // NewListTradesService init listing trades service
	// NewListTradesServiceDo()

	// // NewHistoricalTradesService init listing trades service
	// NewHistoricalTradesServiceDo()

	// // NewListDepositsService init listing deposits service
	// NewListDepositsServiceDo()

	// // NewGetDepositAddressService init getting deposit address service
	// NewGetDepositAddressServiceDo()

	// // NewCreateWithdrawService init creating withdraw service
	// NewCreateWithdrawServiceDo()

	// // NewListWithdrawsService init listing withdraw service
	// NewListWithdrawsServiceDo()

	// // NewStartUserStreamService init starting user stream service
	// NewStartUserStreamServiceDo()

	// // NewKeepaliveUserStreamService init keep alive user stream service
	// NewKeepaliveUserStreamServiceDo()

	// // NewCloseUserStreamService init closing user stream service
	// NewCloseUserStreamServiceDo()

	// // NewExchangeInfoService init exchange info service
	// NewExchangeInfoServiceDo()

	// // NewRateLimitService init rate limit service
	// NewRateLimitServiceDo()

	// // NewGetAssetDetailService init get asset detail service
	// NewGetAssetDetailServiceDo()

	// // NewAveragePriceService init average price service
	// NewAveragePriceServiceDo()

	// // NewMarginTransferService init margin account transfer service
	// NewMarginTransferServiceDo()

	// // NewMarginLoanService init margin account loan service
	// NewMarginLoanServiceDo()

	// // NewMarginRepayService init margin account repay service
	// NewMarginRepayServiceDo()

	// // NewCreateMarginOrderService init creating margin order service
	// NewCreateMarginOrderServiceDo()

	// // NewCancelMarginOrderService init cancel order service
	// NewCancelMarginOrderServiceDo()

	// // NewCreateMarginOCOService init creating margin order service
	// NewCreateMarginOCOServiceDo()

	// // NewCancelMarginOCOService init cancel order service
	// NewCancelMarginOCOServiceDo()

	// // NewGetMarginOrderService init get order service
	// NewGetMarginOrderServiceDo()

	// // NewListMarginLoansService init list margin loan service
	// NewListMarginLoansServiceDo()

	// // NewListMarginRepaysService init list margin repay service
	// NewListMarginRepaysServiceDo()

	// // NewGetMarginAccountService init get margin account service
	// NewGetMarginAccountServiceDo()

	// // NewGetIsolatedMarginAccountService init get isolated margin asset service
	// NewGetIsolatedMarginAccountServiceDo()

	// NewIsolatedMarginTransferServiceDo()

	// // NewGetMarginAssetService init get margin asset service
	// NewGetMarginAssetServiceDo()

	// // NewGetMarginPairService init get margin pair service
	// NewGetMarginPairServiceDo()

	// // NewGetMarginAllPairsService init get margin all pairs service
	// NewGetMarginAllPairsServiceDo()

	// // NewGetMarginPriceIndexService init get margin price index service
	// NewGetMarginPriceIndexServiceDo()

	// // NewListMarginOpenOrdersService init list margin open orders service
	// NewListMarginOpenOrdersServiceDo()

	// // NewListMarginOrdersService init list margin all orders service
	// NewListMarginOrdersServiceDo()

	// // NewListMarginTradesService init list margin trades service
	// NewListMarginTradesServiceDo()

	// // NewGetMaxBorrowableService init get max borrowable service
	// NewGetMaxBorrowableServiceDo()

	// // NewGetMaxTransferableService init get max transferable service
	// NewGetMaxTransferableServiceDo()

	// // NewStartMarginUserStreamService init starting margin user stream service
	// NewStartMarginUserStreamServiceDo()

	// // NewKeepaliveMarginUserStreamService init keep alive margin user stream service
	// NewKeepaliveMarginUserStreamServiceDo()

	// // NewCloseMarginUserStreamService init closing margin user stream service
	// NewCloseMarginUserStreamServiceDo()

	// // NewStartIsolatedMarginUserStreamService init starting margin user stream service
	// NewStartIsolatedMarginUserStreamServiceDo()

	// // NewKeepaliveIsolatedMarginUserStreamService init keep alive margin user stream service
	// NewKeepaliveIsolatedMarginUserStreamServiceDo()

	// // NewCloseIsolatedMarginUserStreamService init closing margin user stream service
	// NewCloseIsolatedMarginUserStreamServiceDo()

	// // NewFuturesTransferService init futures transfer service
	// NewFuturesTransferServiceDo()

	// // NewListFuturesTransferService init list futures transfer service
	// NewListFuturesTransferServiceDo()

	// // NewListDustLogService init list dust log service
	// NewListDustLogServiceDo()

	// // NewDustTransferService init dust transfer service
	// NewDustTransferServiceDo()

	// // NewListDustService init dust list service
	// NewListDustServiceDo()

	// // NewTransferToSubAccountService transfer to subaccount service
	// NewTransferToSubAccountServiceDo()

	// // NewSubaccountAssetsService init list subaccount assets
	// NewSubaccountAssetsServiceDo()

	// // NewSubaccountSpotSummaryService init subaccount spot summary
	// NewSubaccountSpotSummaryServiceDo()

	// // NewSubaccountDepositAddressService init subaccount deposit address service
	// NewSubaccountDepositAddressServiceDo()

	// // NewAssetDividendService init the asset dividend list service
	// NewAssetDividendServiceDo()

	// // NewUserUniversalTransferService
	// NewUserUniversalTransferServiceDo()

	// // NewAllCoinsInformation
	// NewGetAllCoinsInfoServiceDo()

	// // NewDustTransferService init Get All Margin Assets service
	// NewGetAllMarginAssetsServiceDo()

	// // NewFiatDepositWithdrawHistoryService init the fiat deposit/withdraw history service
	// NewFiatDepositWithdrawHistoryServiceDo()

	// // NewFiatPaymentsHistoryService init the fiat payments history service
	// NewFiatPaymentsHistoryServiceDo()

	// // NewPayTransactionService init the pay transaction service
	// NewPayTradeHistoryServiceDo()

	// // NewFiatPaymentsHistoryService init the spot rebate history service
	// NewSpotRebateHistoryServiceDo()

	// // NewConvertTradeHistoryService init the convert trade history service
	// NewConvertTradeHistoryServiceDo()

	// // NewGetIsolatedMarginAllPairsService init get isolated margin all pairs service
	// NewGetIsolatedMarginAllPairsServiceDo()

	// // NewInterestHistoryService init the interest history service
	// NewInterestHistoryServiceDo()

	// // NewTradeFeeService init the trade fee service
	// NewTradeFeeServiceDo()

	// // NewC2CTradeHistoryService init the c2c trade history service
	// NewC2CTradeHistoryServiceDo()

	// // NewStakingProductPositionService init the staking product position service
	// NewStakingProductPositionServiceDo()

	// // NewStakingHistoryService init the staking history service
	// NewStakingHistoryServiceDo()

	// // NewGetAllLiquidityPoolService init the get all swap pool service
	// NewGetAllLiquidityPoolServiceDo()

	// // NewGetLiquidityPoolDetailService init the get liquidity pool detial service
	// NewGetLiquidityPoolDetailServiceDo()

	// // NewAddLiquidityPreviewService init the add liquidity preview service
	// NewAddLiquidityPreviewServiceDo()

	// // NewGetSwapQuoteService init the add liquidity preview service
	// NewGetSwapQuoteServiceDo()

	// // NewSwapService init the swap service
	// NewSwapServiceDo()

	// // NewAddLiquidityService init the add liquidity service
	// NewAddLiquidityServiceDo()

	// // NewGetUserSwapRecordsService init the service for listing the swap records
	// NewGetUserSwapRecordsServiceDo()

	// // NewClaimRewardService init the service for liquidity pool rewarding
	// NewClaimRewardServiceDo()

	// // NewRemoveLiquidityService init the service to remvoe liquidity
	// NewRemoveLiquidityServiceDo()

	// // NewQueryClaimedRewardHistoryService init the service to query reward claiming history
	// NewQueryClaimedRewardHistoryServiceDo()

	// // NewGetBNBBurnService init the service to get BNB Burn on spot trade and margin interest
	// NewGetBNBBurnServiceDo()

	// // NewToggleBNBBurnService init the service to toggle BNB Burn on spot trade and margin interest
	// NewToggleBNBBurnServiceDo()

	// // NewInternalUniversalTransferService Universal Transfer (For Master Account)
	// NewInternalUniversalTransferServiceDo()

	// // NewInternalUniversalTransferHistoryService Query Universal Transfer History (For Master Account)
	// NewInternalUniversalTransferHistoryServiceDo()

	// // NewSubAccountListService Query Sub-account List (For Master Account)
	// NewSubAccountListServiceDo()

	// // NewGetUserAsset Get user assets, just for positive data
	// NewGetUserAssetDo()

	// // NewManagedSubAccountDepositService Deposit Assets Into The Managed Sub-account（For Investor Master Account）
	// NewManagedSubAccountDepositServiceDo()

	// // NewManagedSubAccountWithdrawalService Withdrawal Assets From The Managed Sub-account（For Investor Master Account）
	// NewManagedSubAccountWithdrawalServiceDo()

	// // NewManagedSubAccountAssetsService Withdrawal Assets From The Managed Sub-account（For Investor Master Account）
	// NewManagedSubAccountAssetsServiceDo()

	// // NewSubAccountFuturesAccountService Get Detail on Sub-account's Futures Account (For Master Account)
	// NewSubAccountFuturesAccountServiceDo()

	// // NewSubAccountFuturesSummaryV1Service Get Summary of Sub-account's Futures Account (For Master Account)
	// NewSubAccountFuturesSummaryV1ServiceDo()

	// // NewSubAccountFuturesTransferV1Service Futures Transfer for Sub-account (For Master Account)
	// NewSubAccountFuturesTransferV1ServiceDo()
}
