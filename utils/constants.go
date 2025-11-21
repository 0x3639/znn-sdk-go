package utils

import (
	"time"

	"github.com/0x3639/znn-sdk-go/embedded"
)

// =============================================================================
// NOM Constants
// =============================================================================

// CoinDecimals is the number of decimals for ZNN and QSR tokens
const CoinDecimals = embedded.CoinDecimals

// OneZnn represents 1 ZNN in base units (10^8)
const OneZnn = embedded.OneZnn

// OneQsr represents 1 QSR in base units (10^8)
const OneQsr = embedded.OneQsr

// IntervalBetweenMomentums is the time between momentum blocks (10 seconds)
const IntervalBetweenMomentums = 10 * time.Second
