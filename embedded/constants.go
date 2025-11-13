package embedded

import (
	"math/big"
	"regexp"
)

// =============================================================================
// Base Constants
// =============================================================================

const (
	// CoinDecimals is the number of decimals for ZNN and QSR tokens
	CoinDecimals = 8

	// OneZnn represents 1 ZNN in base units (10^8)
	OneZnn = 100000000

	// OneQsr represents 1 QSR in base units (10^8)
	OneQsr = 100000000
)

// GenesisTimestamp is the Unix timestamp of the genesis block
const GenesisTimestamp = 1637755200

// =============================================================================
// Plasma Constants
// =============================================================================

var (
	// FuseMinQsrAmount is the minimum QSR amount required to fuse plasma
	FuseMinQsrAmount = big.NewInt(10 * OneQsr)

	// MinPlasmaAmount is the minimum plasma amount
	MinPlasmaAmount = big.NewInt(21000)
)

// =============================================================================
// Pillar Constants
// =============================================================================

var (
	// MinDelegationAmount is the minimum ZNN amount for delegation
	MinDelegationAmount = big.NewInt(1 * OneZnn)

	// PillarRegisterZnnAmount is the ZNN amount required to register a pillar
	PillarRegisterZnnAmount = big.NewInt(15000 * OneZnn)

	// PillarRegisterQsrAmount is the QSR amount required to register a pillar
	PillarRegisterQsrAmount = big.NewInt(150000 * OneQsr)
)

const (
	// PillarNameMaxLength is the maximum length for pillar names
	PillarNameMaxLength = 40
)

var (
	// PillarNameRegExp is the regex pattern for valid pillar names
	PillarNameRegExp = regexp.MustCompile(`^([a-zA-Z0-9]+[-._]?)*[a-zA-Z0-9]$`)
)

// =============================================================================
// Sentinel Constants
// =============================================================================

var (
	// SentinelRegisterZnnAmount is the ZNN amount required to register a sentinel
	SentinelRegisterZnnAmount = big.NewInt(5000 * OneZnn)

	// SentinelRegisterQsrAmount is the QSR amount required to register a sentinel
	SentinelRegisterQsrAmount = big.NewInt(50000 * OneQsr)
)

// =============================================================================
// Staking Constants
// =============================================================================

var (
	// StakeMinZnnAmount is the minimum ZNN amount for staking
	StakeMinZnnAmount = big.NewInt(1 * OneZnn)
)

const (
	// StakeTimeUnitSec is the time unit for staking (30 days in seconds)
	StakeTimeUnitSec = 30 * 24 * 60 * 60

	// StakeTimeMinSec is the minimum staking duration in seconds (1 month)
	StakeTimeMinSec = 1 * StakeTimeUnitSec

	// StakeTimeMaxSec is the maximum staking duration in seconds (12 months)
	StakeTimeMaxSec = 12 * StakeTimeUnitSec

	// StakeUnitDurationName is the human-readable name for the staking time unit
	StakeUnitDurationName = "month"
)

// =============================================================================
// Token Constants
// =============================================================================

var (
	// TokenZtsIssueFeeInZnn is the ZNN fee for issuing a new token
	TokenZtsIssueFeeInZnn = big.NewInt(1 * OneZnn)

	// MinTokenTotalMaxSupply is the minimum value for token total/max supply
	MinTokenTotalMaxSupply = big.NewInt(1)

	// BigP255 represents 2^255
	BigP255 = new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil)

	// BigP255m1 represents 2^255 - 1 (maximum token supply)
	BigP255m1 = new(big.Int).Sub(BigP255, big.NewInt(1))
)

const (
	// TokenNameMaxLength is the maximum length for token names
	TokenNameMaxLength = 40

	// TokenSymbolMaxLength is the maximum length for token symbols
	TokenSymbolMaxLength = 10
)

var (
	// TokenSymbolExceptions is the list of reserved token symbols
	TokenSymbolExceptions = []string{"ZNN", "QSR"}

	// TokenNameRegExp is the regex pattern for valid token names
	TokenNameRegExp = regexp.MustCompile(`^([a-zA-Z0-9]+[-._]?)*[a-zA-Z0-9]$`)

	// TokenSymbolRegExp is the regex pattern for valid token symbols
	TokenSymbolRegExp = regexp.MustCompile(`^[A-Z0-9]+$`)

	// TokenDomainRegExp is the regex pattern for valid token domains
	TokenDomainRegExp = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9-]{0,61}[A-Za-z0-9]\.)+[A-Za-z]{2,}$`)
)

// =============================================================================
// Accelerator Constants
// =============================================================================

var (
	// ProjectCreationFeeInZnn is the ZNN fee for creating an accelerator project
	ProjectCreationFeeInZnn = big.NewInt(1 * OneZnn)

	// ZnnProjectMaximumFunds is the maximum ZNN that can be requested per project
	ZnnProjectMaximumFunds = big.NewInt(5000 * OneZnn)

	// QsrProjectMaximumFunds is the maximum QSR that can be requested per project
	QsrProjectMaximumFunds = big.NewInt(50000 * OneQsr)

	// ZnnProjectMinimumFunds is the minimum ZNN that can be requested per project
	ZnnProjectMinimumFunds = big.NewInt(10 * OneZnn)

	// QsrProjectMinimumFunds is the minimum QSR that can be requested per project
	QsrProjectMinimumFunds = big.NewInt(100 * OneQsr)
)

const (
	// ProjectDescriptionMaxLength is the maximum length for project descriptions
	ProjectDescriptionMaxLength = 240

	// ProjectNameMaxLength is the maximum length for project names
	ProjectNameMaxLength = 30

	// ProjectVotingStatus indicates a project is in voting phase
	ProjectVotingStatus = 0

	// ProjectActiveStatus indicates a project is active
	ProjectActiveStatus = 1

	// ProjectPaidStatus indicates a project has been paid
	ProjectPaidStatus = 2

	// ProjectClosedStatus indicates a project is closed
	ProjectClosedStatus = 3
)

var (
	// ProjectUrlRegExp is the regex pattern for valid project URLs
	ProjectUrlRegExp = regexp.MustCompile(`^[a-zA-Z0-9]{2,60}\.[a-zA-Z]{1,6}([a-zA-Z0-9()@:%_\\+.~#?&/=-]{0,100})$`)
)

// =============================================================================
// Swap Constants
// =============================================================================

const (
	// SwapAssetDecayTimestampStart is the start timestamp for swap asset decay
	SwapAssetDecayTimestampStart = 1645531200

	// SwapAssetDecayEpochsOffset is the epoch offset for swap asset decay
	SwapAssetDecayEpochsOffset = 30 * 3

	// SwapAssetDecayTickEpochs is the number of epochs per decay tick
	SwapAssetDecayTickEpochs = 30

	// SwapAssetDecayTickValuePercentage is the percentage decay per tick
	SwapAssetDecayTickValuePercentage = 10
)

// =============================================================================
// Spork Constants
// =============================================================================

const (
	// SporkNameMinLength is the minimum length for spork names
	SporkNameMinLength = 5

	// SporkNameMaxLength is the maximum length for spork names
	SporkNameMaxLength = 40

	// SporkDescriptionMaxLength is the maximum length for spork descriptions
	SporkDescriptionMaxLength = 400
)

// =============================================================================
// HTLC Constants
// =============================================================================

const (
	// HtlcPreimageMinLength is the minimum length for HTLC preimages
	HtlcPreimageMinLength = 1

	// HtlcPreimageMaxLength is the maximum length for HTLC preimages
	HtlcPreimageMaxLength = 255

	// HtlcPreimageDefaultLength is the default length for HTLC preimages
	HtlcPreimageDefaultLength = 32

	// HtlcHashTypeSha3 indicates SHA3-256 hash type
	HtlcHashTypeSha3 = 0

	// HtlcHashTypeSha256 indicates SHA-256 hash type
	HtlcHashTypeSha256 = 1
)

// =============================================================================
// Bridge Constants
// =============================================================================

const (
	// BridgeMinGuardians is the minimum number of bridge guardians
	BridgeMinGuardians = 5

	// BridgeMaximumFee is the maximum bridge fee (in basis points)
	BridgeMaximumFee = 10000
)
