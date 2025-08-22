package testutil

import (
	"os"
	"testing"

	binance "github.com/adshao/go-binance/v2"
	futures "github.com/adshao/go-binance/v2/futures"
)

// readSpotKeys returns SPOT_* test keys or falls back to generic API_KEY/SECRET_KEY.
func readSpotKeys() (api, sec string) {
	api = os.Getenv("SPOT_TEST_BINANCE_API_KEY")
	sec = os.Getenv("SPOT_TEST_BINANCE_SECRET_KEY")
	if api == "" || sec == "" {
		api = os.Getenv("API_KEY")
		sec = os.Getenv("SECRET_KEY")
	}
	return
}

// readFuturesKeys returns FUTURE_* test keys or falls back to generic API_KEY/SECRET_KEY.
func readFuturesKeys() (api, sec string) {
	api = os.Getenv("FUTURE_TEST_BINANCE_API_KEY")
	sec = os.Getenv("FUTURE_TEST_BINANCE_SECRET_KEY")
	if api == "" || sec == "" {
		api = os.Getenv("API_KEY")
		sec = os.Getenv("SECRET_KEY")
	}
	return
}

// SpotClient creates a Spot client or skips the test if keys are missing.
// useTestnet controls binance.UseTestnet.
func SpotClientTB(t testing.TB, useTestnet bool) *binance.Client {
	t.Helper()
	api, sec := readSpotKeys()
	if api == "" || sec == "" {
		t.Skip("Пропущено: немає ключів для Spot (SPOT_TEST_BINANCE_API_KEY/SPOT_TEST_BINANCE_SECRET_KEY або API_KEY/SECRET_KEY)")
	}
	binance.UseTestnet = useTestnet
	return binance.NewClient(api, sec)
}

// FuturesClient creates a Futures client or skips the test if keys are missing.
// useTestnet controls futures.UseTestnet.
func FuturesClientTB(t testing.TB, useTestnet bool) *futures.Client {
	t.Helper()
	api, sec := readFuturesKeys()
	if api == "" || sec == "" {
		t.Skip("Пропущено: немає ключів для Futures (FUTURE_TEST_BINANCE_API_KEY/FUTURE_TEST_BINANCE_SECRET_KEY або API_KEY/SECRET_KEY)")
	}
	futures.UseTestnet = useTestnet
	return futures.NewClient(api, sec)
}

// Convenience wrappers defaulting to testnet=true.
func SpotClient(t testing.TB) *binance.Client    { return SpotClientTB(t, true) }
func FuturesClient(t testing.TB) *futures.Client { return FuturesClientTB(t, true) }
