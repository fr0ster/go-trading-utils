# go-trading-utils Release Notes

## v0.2.13 - 2025-08-26

Changes:
- 767774ff feat: add account event types and structures for trading operations

## v0.2.12 - 2025-08-22

Changes:
- Upgrade dependencies: go-binance v2.8.5, golang.org/x/sys v0.35.0, and others.
- Make tests parallel-safe with t.Parallel across packages.
- Introduce internal testutil to standardize API client creation with env-based skips.
- Refactor tests to use internal/testutil (SpotClient/FuturesClient) and avoid global UseTestnet toggles.
- Normalize env var behavior: SPOT_TEST_* and FUTURE_TEST_* with API_KEY/SECRET_KEY fallback.

Notes:
- Network-dependent tests are skipped if API keys are not provided via environment variables.
- No breaking API changes to library packages.
