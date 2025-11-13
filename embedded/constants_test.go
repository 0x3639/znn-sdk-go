package embedded

import (
	"math/big"
	"testing"
)

// =============================================================================
// Base Constants Tests
// =============================================================================

func TestCoinDecimals(t *testing.T) {
	if CoinDecimals != 8 {
		t.Errorf("CoinDecimals = %d, want 8", CoinDecimals)
	}
}

func TestOneZnn(t *testing.T) {
	if OneZnn != 100000000 {
		t.Errorf("OneZnn = %d, want 100000000", OneZnn)
	}
}

func TestOneQsr(t *testing.T) {
	if OneQsr != 100000000 {
		t.Errorf("OneQsr = %d, want 100000000", OneQsr)
	}
}

func TestGenesisTimestamp(t *testing.T) {
	if GenesisTimestamp != 1637755200 {
		t.Errorf("GenesisTimestamp = %d, want 1637755200", GenesisTimestamp)
	}
}

// =============================================================================
// Plasma Constants Tests
// =============================================================================

func TestFuseMinQsrAmount(t *testing.T) {
	expected := big.NewInt(10 * OneQsr)
	if FuseMinQsrAmount.Cmp(expected) != 0 {
		t.Errorf("FuseMinQsrAmount = %s, want %s", FuseMinQsrAmount, expected)
	}
}

func TestMinPlasmaAmount(t *testing.T) {
	expected := big.NewInt(21000)
	if MinPlasmaAmount.Cmp(expected) != 0 {
		t.Errorf("MinPlasmaAmount = %s, want %s", MinPlasmaAmount, expected)
	}
}

// =============================================================================
// Pillar Constants Tests
// =============================================================================

func TestMinDelegationAmount(t *testing.T) {
	expected := big.NewInt(1 * OneZnn)
	if MinDelegationAmount.Cmp(expected) != 0 {
		t.Errorf("MinDelegationAmount = %s, want %s", MinDelegationAmount, expected)
	}
}

func TestPillarRegisterZnnAmount(t *testing.T) {
	expected := big.NewInt(15000 * OneZnn)
	if PillarRegisterZnnAmount.Cmp(expected) != 0 {
		t.Errorf("PillarRegisterZnnAmount = %s, want %s", PillarRegisterZnnAmount, expected)
	}
}

func TestPillarRegisterQsrAmount(t *testing.T) {
	expected := big.NewInt(150000 * OneQsr)
	if PillarRegisterQsrAmount.Cmp(expected) != 0 {
		t.Errorf("PillarRegisterQsrAmount = %s, want %s", PillarRegisterQsrAmount, expected)
	}
}

func TestPillarNameMaxLength(t *testing.T) {
	if PillarNameMaxLength != 40 {
		t.Errorf("PillarNameMaxLength = %d, want 40", PillarNameMaxLength)
	}
}

func TestPillarNameRegExp(t *testing.T) {
	validNames := []string{
		"MyPillar",
		"pillar-1",
		"my.pillar",
		"test_pillar",
		"Pillar123",
	}

	for _, name := range validNames {
		if !PillarNameRegExp.MatchString(name) {
			t.Errorf("PillarNameRegExp should match %s", name)
		}
	}

	invalidNames := []string{
		"-invalid",
		"invalid-",
		".invalid",
		"invalid.",
		"invalid--name",
		"",
	}

	for _, name := range invalidNames {
		if PillarNameRegExp.MatchString(name) {
			t.Errorf("PillarNameRegExp should not match %s", name)
		}
	}
}

// =============================================================================
// Sentinel Constants Tests
// =============================================================================

func TestSentinelRegisterZnnAmount(t *testing.T) {
	expected := big.NewInt(5000 * OneZnn)
	if SentinelRegisterZnnAmount.Cmp(expected) != 0 {
		t.Errorf("SentinelRegisterZnnAmount = %s, want %s", SentinelRegisterZnnAmount, expected)
	}
}

func TestSentinelRegisterQsrAmount(t *testing.T) {
	expected := big.NewInt(50000 * OneQsr)
	if SentinelRegisterQsrAmount.Cmp(expected) != 0 {
		t.Errorf("SentinelRegisterQsrAmount = %s, want %s", SentinelRegisterQsrAmount, expected)
	}
}

// =============================================================================
// Staking Constants Tests
// =============================================================================

func TestStakeMinZnnAmount(t *testing.T) {
	expected := big.NewInt(1 * OneZnn)
	if StakeMinZnnAmount.Cmp(expected) != 0 {
		t.Errorf("StakeMinZnnAmount = %s, want %s", StakeMinZnnAmount, expected)
	}
}

func TestStakeTimeConstants(t *testing.T) {
	if StakeTimeUnitSec != 30*24*60*60 {
		t.Errorf("StakeTimeUnitSec = %d, want %d", StakeTimeUnitSec, 30*24*60*60)
	}

	if StakeTimeMinSec != StakeTimeUnitSec {
		t.Errorf("StakeTimeMinSec = %d, want %d", StakeTimeMinSec, StakeTimeUnitSec)
	}

	if StakeTimeMaxSec != 12*StakeTimeUnitSec {
		t.Errorf("StakeTimeMaxSec = %d, want %d", StakeTimeMaxSec, 12*StakeTimeUnitSec)
	}

	if StakeUnitDurationName != "month" {
		t.Errorf("StakeUnitDurationName = %s, want month", StakeUnitDurationName)
	}
}

// =============================================================================
// Token Constants Tests
// =============================================================================

func TestTokenZtsIssueFeeInZnn(t *testing.T) {
	expected := big.NewInt(1 * OneZnn)
	if TokenZtsIssueFeeInZnn.Cmp(expected) != 0 {
		t.Errorf("TokenZtsIssueFeeInZnn = %s, want %s", TokenZtsIssueFeeInZnn, expected)
	}
}

func TestMinTokenTotalMaxSupply(t *testing.T) {
	expected := big.NewInt(1)
	if MinTokenTotalMaxSupply.Cmp(expected) != 0 {
		t.Errorf("MinTokenTotalMaxSupply = %s, want %s", MinTokenTotalMaxSupply, expected)
	}
}

func TestBigP255(t *testing.T) {
	// 2^255
	expected := new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil)
	if BigP255.Cmp(expected) != 0 {
		t.Errorf("BigP255 value mismatch")
	}
}

func TestBigP255m1(t *testing.T) {
	// 2^255 - 1
	expected := new(big.Int).Sub(BigP255, big.NewInt(1))
	if BigP255m1.Cmp(expected) != 0 {
		t.Errorf("BigP255m1 value mismatch")
	}
}

func TestTokenNameMaxLength(t *testing.T) {
	if TokenNameMaxLength != 40 {
		t.Errorf("TokenNameMaxLength = %d, want 40", TokenNameMaxLength)
	}
}

func TestTokenSymbolMaxLength(t *testing.T) {
	if TokenSymbolMaxLength != 10 {
		t.Errorf("TokenSymbolMaxLength = %d, want 10", TokenSymbolMaxLength)
	}
}

func TestTokenSymbolExceptions(t *testing.T) {
	if len(TokenSymbolExceptions) != 2 {
		t.Errorf("len(TokenSymbolExceptions) = %d, want 2", len(TokenSymbolExceptions))
	}

	if TokenSymbolExceptions[0] != "ZNN" {
		t.Errorf("TokenSymbolExceptions[0] = %s, want ZNN", TokenSymbolExceptions[0])
	}

	if TokenSymbolExceptions[1] != "QSR" {
		t.Errorf("TokenSymbolExceptions[1] = %s, want QSR", TokenSymbolExceptions[1])
	}
}

func TestTokenNameRegExp(t *testing.T) {
	validNames := []string{
		"MyToken",
		"token-1",
		"my.token",
		"test_token",
		"Token123",
	}

	for _, name := range validNames {
		if !TokenNameRegExp.MatchString(name) {
			t.Errorf("TokenNameRegExp should match %s", name)
		}
	}

	invalidNames := []string{
		"-invalid",
		"invalid-",
		".invalid",
		"invalid.",
		"",
	}

	for _, name := range invalidNames {
		if TokenNameRegExp.MatchString(name) {
			t.Errorf("TokenNameRegExp should not match %s", name)
		}
	}
}

func TestTokenSymbolRegExp(t *testing.T) {
	validSymbols := []string{
		"ZNN",
		"QSR",
		"TEST",
		"TOKEN123",
		"ABC",
	}

	for _, symbol := range validSymbols {
		if !TokenSymbolRegExp.MatchString(symbol) {
			t.Errorf("TokenSymbolRegExp should match %s", symbol)
		}
	}

	invalidSymbols := []string{
		"znn",
		"Test",
		"token-1",
		"",
		"abc",
	}

	for _, symbol := range invalidSymbols {
		if TokenSymbolRegExp.MatchString(symbol) {
			t.Errorf("TokenSymbolRegExp should not match %s", symbol)
		}
	}
}

func TestTokenDomainRegExp(t *testing.T) {
	validDomains := []string{
		"example.com",
		"test.example.com",
		"my-domain.io",
		"test123.org",
	}

	for _, domain := range validDomains {
		if !TokenDomainRegExp.MatchString(domain) {
			t.Errorf("TokenDomainRegExp should match %s", domain)
		}
	}

	invalidDomains := []string{
		"invalid",
		"-invalid.com",
		"invalid-.com",
		".com",
		"",
	}

	for _, domain := range invalidDomains {
		if TokenDomainRegExp.MatchString(domain) {
			t.Errorf("TokenDomainRegExp should not match %s", domain)
		}
	}
}

// =============================================================================
// Accelerator Constants Tests
// =============================================================================

func TestProjectCreationFeeInZnn(t *testing.T) {
	expected := big.NewInt(1 * OneZnn)
	if ProjectCreationFeeInZnn.Cmp(expected) != 0 {
		t.Errorf("ProjectCreationFeeInZnn = %s, want %s", ProjectCreationFeeInZnn, expected)
	}
}

func TestAcceleratorFundLimits(t *testing.T) {
	if ZnnProjectMaximumFunds.Cmp(big.NewInt(5000*OneZnn)) != 0 {
		t.Error("ZnnProjectMaximumFunds mismatch")
	}

	if QsrProjectMaximumFunds.Cmp(big.NewInt(50000*OneQsr)) != 0 {
		t.Error("QsrProjectMaximumFunds mismatch")
	}

	if ZnnProjectMinimumFunds.Cmp(big.NewInt(10*OneZnn)) != 0 {
		t.Error("ZnnProjectMinimumFunds mismatch")
	}

	if QsrProjectMinimumFunds.Cmp(big.NewInt(100*OneQsr)) != 0 {
		t.Error("QsrProjectMinimumFunds mismatch")
	}
}

func TestProjectConstants(t *testing.T) {
	if ProjectDescriptionMaxLength != 240 {
		t.Errorf("ProjectDescriptionMaxLength = %d, want 240", ProjectDescriptionMaxLength)
	}

	if ProjectNameMaxLength != 30 {
		t.Errorf("ProjectNameMaxLength = %d, want 30", ProjectNameMaxLength)
	}

	if ProjectVotingStatus != 0 {
		t.Errorf("ProjectVotingStatus = %d, want 0", ProjectVotingStatus)
	}

	if ProjectActiveStatus != 1 {
		t.Errorf("ProjectActiveStatus = %d, want 1", ProjectActiveStatus)
	}

	if ProjectPaidStatus != 2 {
		t.Errorf("ProjectPaidStatus = %d, want 2", ProjectPaidStatus)
	}

	if ProjectClosedStatus != 3 {
		t.Errorf("ProjectClosedStatus = %d, want 3", ProjectClosedStatus)
	}
}

// =============================================================================
// Swap Constants Tests
// =============================================================================

func TestSwapConstants(t *testing.T) {
	if SwapAssetDecayTimestampStart != 1645531200 {
		t.Errorf("SwapAssetDecayTimestampStart = %d, want 1645531200", SwapAssetDecayTimestampStart)
	}

	if SwapAssetDecayEpochsOffset != 90 {
		t.Errorf("SwapAssetDecayEpochsOffset = %d, want 90", SwapAssetDecayEpochsOffset)
	}

	if SwapAssetDecayTickEpochs != 30 {
		t.Errorf("SwapAssetDecayTickEpochs = %d, want 30", SwapAssetDecayTickEpochs)
	}

	if SwapAssetDecayTickValuePercentage != 10 {
		t.Errorf("SwapAssetDecayTickValuePercentage = %d, want 10", SwapAssetDecayTickValuePercentage)
	}
}

// =============================================================================
// Spork Constants Tests
// =============================================================================

func TestSporkConstants(t *testing.T) {
	if SporkNameMinLength != 5 {
		t.Errorf("SporkNameMinLength = %d, want 5", SporkNameMinLength)
	}

	if SporkNameMaxLength != 40 {
		t.Errorf("SporkNameMaxLength = %d, want 40", SporkNameMaxLength)
	}

	if SporkDescriptionMaxLength != 400 {
		t.Errorf("SporkDescriptionMaxLength = %d, want 400", SporkDescriptionMaxLength)
	}
}

// =============================================================================
// HTLC Constants Tests
// =============================================================================

func TestHtlcConstants(t *testing.T) {
	if HtlcPreimageMinLength != 1 {
		t.Errorf("HtlcPreimageMinLength = %d, want 1", HtlcPreimageMinLength)
	}

	if HtlcPreimageMaxLength != 255 {
		t.Errorf("HtlcPreimageMaxLength = %d, want 255", HtlcPreimageMaxLength)
	}

	if HtlcPreimageDefaultLength != 32 {
		t.Errorf("HtlcPreimageDefaultLength = %d, want 32", HtlcPreimageDefaultLength)
	}

	if HtlcHashTypeSha3 != 0 {
		t.Errorf("HtlcHashTypeSha3 = %d, want 0", HtlcHashTypeSha3)
	}

	if HtlcHashTypeSha256 != 1 {
		t.Errorf("HtlcHashTypeSha256 = %d, want 1", HtlcHashTypeSha256)
	}
}

// =============================================================================
// Bridge Constants Tests
// =============================================================================

func TestBridgeConstants(t *testing.T) {
	if BridgeMinGuardians != 5 {
		t.Errorf("BridgeMinGuardians = %d, want 5", BridgeMinGuardians)
	}

	if BridgeMaximumFee != 10000 {
		t.Errorf("BridgeMaximumFee = %d, want 10000", BridgeMaximumFee)
	}
}
