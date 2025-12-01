package embedded

import (
	"encoding/json"
	"math/big"

	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
)

// BridgeInfo represents bridge contract configuration.
//
// The bridge enables cross-chain transfers between Zenon and other networks
// (e.g., Ethereum). This type contains the current configuration and state
// of the bridge contract.
//
// Fields:
//   - Administrator: Address that can configure the bridge
//   - CompressedTssECDSAPubKey: Compressed TSS public key for signing
//   - DecompressedTssECDSAPubKey: Decompressed TSS public key
//   - AllowKeyGen: Whether key generation ceremonies are enabled
//   - Halted: Whether bridge operations are currently paused
//   - UnhaltedAt: Momentum height when the bridge was last unhalted
//   - UnhaltDurationInMomentums: How long the bridge stays unhalted
//   - TssNonce: Current nonce for TSS operations
//   - Metadata: Additional bridge configuration data
type BridgeInfo struct {
	Administrator              types.Address `json:"administrator"`
	CompressedTssECDSAPubKey   string        `json:"compressedTssECDSAPubKey"`
	DecompressedTssECDSAPubKey string        `json:"decompressedTssECDSAPubKey"`
	AllowKeyGen                bool          `json:"allowKeyGen"`
	Halted                     bool          `json:"halted"`
	UnhaltedAt                 uint64        `json:"unhaltedAt"`
	UnhaltDurationInMomentums  uint64        `json:"unhaltDurationInMomentums"`
	TssNonce                   uint64        `json:"tssNonce"`
	Metadata                   string        `json:"metadata"`
}

// OrchestratorInfo represents orchestrator configuration.
//
// Orchestrators are the bridge operators that facilitate cross-chain transfers
// using threshold signature scheme (TSS). This type contains configuration
// parameters for orchestrator operations.
//
// Fields:
//   - WindowSize: Time window for orchestrator operations
//   - KeyGenThreshold: Minimum participants required for key generation
//   - ConfirmationsToFinality: Block confirmations needed on external chain
//   - EstimatedMomentumTime: Estimated time per momentum in seconds
//   - AllowKeyGenHeight: Momentum height when key generation becomes allowed
type OrchestratorInfo struct {
	WindowSize              uint64 `json:"windowSize"`
	KeyGenThreshold         uint32 `json:"keyGenThreshold"`
	ConfirmationsToFinality uint32 `json:"confirmationsToFinality"`
	EstimatedMomentumTime   uint32 `json:"estimatedMomentumTime"`
	AllowKeyGenHeight       uint64 `json:"allowKeyGenHeight"`
}

// TokenPair represents a bridge token pair configuration.
//
// This type defines the mapping between a Zenon ZTS token and its corresponding
// token on an external chain, along with bridge operation parameters.
//
// Fields:
//   - TokenStandard: ZTS identifier on the Zenon side
//   - TokenAddress: Contract address on the external chain
//   - Bridgeable: Whether tokens can be wrapped (Zenon -> external)
//   - Redeemable: Whether tokens can be unwrapped (external -> Zenon)
//   - Owned: Whether the bridge owns the external token contract
//   - MinAmount: Minimum amount for bridge operations (in base units, 8 decimals)
//   - FeePercentage: Fee charged for bridge operations (basis points)
//   - RedeemDelay: Delay before unwrapped tokens can be claimed
//   - Metadata: Additional token pair configuration data
type TokenPair struct {
	TokenStandard types.ZenonTokenStandard `json:"tokenStandard"`
	TokenAddress  string                   `json:"tokenAddress"`
	Bridgeable    bool                     `json:"bridgeable"`
	Redeemable    bool                     `json:"redeemable"`
	Owned         bool                     `json:"owned"`
	MinAmount     *big.Int                 `json:"minAmount"`
	FeePercentage uint32                   `json:"feePercentage"`
	RedeemDelay   uint32                   `json:"redeemDelay"`
	Metadata      string                   `json:"metadata"`
}

// tokenPairJSON is used for JSON unmarshaling with string amounts
type tokenPairJSON struct {
	TokenStandard types.ZenonTokenStandard `json:"tokenStandard"`
	TokenAddress  string                   `json:"tokenAddress"`
	Bridgeable    bool                     `json:"bridgeable"`
	Redeemable    bool                     `json:"redeemable"`
	Owned         bool                     `json:"owned"`
	MinAmount     string                   `json:"minAmount"`
	FeePercentage uint32                   `json:"feePercentage"`
	RedeemDelay   uint32                   `json:"redeemDelay"`
	Metadata      string                   `json:"metadata"`
}

func (t *TokenPair) UnmarshalJSON(data []byte) error {
	var aux tokenPairJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	t.TokenStandard = aux.TokenStandard
	t.TokenAddress = aux.TokenAddress
	t.Bridgeable = aux.Bridgeable
	t.Redeemable = aux.Redeemable
	t.Owned = aux.Owned
	t.MinAmount = common.StringToBigInt(aux.MinAmount)
	t.FeePercentage = aux.FeePercentage
	t.RedeemDelay = aux.RedeemDelay
	t.Metadata = aux.Metadata
	return nil
}

// BridgeNetworkInfo represents bridge network configuration.
//
// This type contains information about an external blockchain network that
// is connected via the bridge, including supported token pairs.
//
// Fields:
//   - NetworkClass: Category of network (e.g., EVM-based)
//   - ChainId: Chain identifier for the external network
//   - Name: Human-readable name of the network
//   - ContractAddress: Bridge contract address on the external network
//   - Metadata: Additional network configuration data
//   - TokenPairs: List of token pairs supported on this network
type BridgeNetworkInfo struct {
	NetworkClass    uint32       `json:"networkClass"`
	ChainId         uint32       `json:"chainId"`
	Name            string       `json:"name"`
	ContractAddress string       `json:"contractAddress"`
	Metadata        string       `json:"metadata"`
	TokenPairs      []*TokenPair `json:"tokenPairs"`
}

// BridgeNetworkInfoList represents a paginated list of bridge networks.
//
// Fields:
//   - Count: Total number of networks matching the query
//   - List: Slice of BridgeNetworkInfo entries for the current page
type BridgeNetworkInfoList struct {
	Count int                  `json:"count"`
	List  []*BridgeNetworkInfo `json:"list"`
}

// WrapTokenRequest represents a wrap token request.
//
// When transferring tokens from Zenon to an external chain, a wrap request is
// created. Orchestrators process these requests and mint tokens on the
// destination chain.
//
// Fields:
//   - NetworkClass: Category of destination network
//   - ChainId: Chain identifier of destination network
//   - Id: Unique identifier for this wrap request
//   - ToAddress: Recipient address on the external chain
//   - TokenStandard: ZTS identifier of the token being wrapped
//   - TokenAddress: Token contract address on the external chain
//   - Amount: Amount to wrap (in base units, 8 decimals)
//   - Fee: Fee charged for the wrap operation (in base units)
//   - Signature: TSS signature for the wrap (added by orchestrators)
//   - CreationMomentumHeight: Momentum height when request was created
//   - ConfirmationsToFinality: Number of confirmations remaining until finality
type WrapTokenRequest struct {
	NetworkClass            uint32                   `json:"networkClass"`
	ChainId                 uint32                   `json:"chainId"`
	Id                      types.Hash               `json:"id"`
	ToAddress               string                   `json:"toAddress"`
	TokenStandard           types.ZenonTokenStandard `json:"tokenStandard"`
	TokenAddress            string                   `json:"tokenAddress"`
	Amount                  *big.Int                 `json:"amount"`
	Fee                     *big.Int                 `json:"fee"`
	Signature               string                   `json:"signature"`
	CreationMomentumHeight  uint64                   `json:"creationMomentumHeight"`
	ConfirmationsToFinality uint32                   `json:"confirmationsToFinality"`
}

// wrapTokenRequestJSON is used for JSON unmarshaling with string amounts
type wrapTokenRequestJSON struct {
	NetworkClass            uint32                   `json:"networkClass"`
	ChainId                 uint32                   `json:"chainId"`
	Id                      types.Hash               `json:"id"`
	ToAddress               string                   `json:"toAddress"`
	TokenStandard           types.ZenonTokenStandard `json:"tokenStandard"`
	TokenAddress            string                   `json:"tokenAddress"`
	Amount                  string                   `json:"amount"`
	Fee                     string                   `json:"fee"`
	Signature               string                   `json:"signature"`
	CreationMomentumHeight  uint64                   `json:"creationMomentumHeight"`
	ConfirmationsToFinality uint32                   `json:"confirmationsToFinality"`
}

func (w *WrapTokenRequest) UnmarshalJSON(data []byte) error {
	var aux wrapTokenRequestJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	w.NetworkClass = aux.NetworkClass
	w.ChainId = aux.ChainId
	w.Id = aux.Id
	w.ToAddress = aux.ToAddress
	w.TokenStandard = aux.TokenStandard
	w.TokenAddress = aux.TokenAddress
	w.Amount = common.StringToBigInt(aux.Amount)
	w.Fee = common.StringToBigInt(aux.Fee)
	w.Signature = aux.Signature
	w.CreationMomentumHeight = aux.CreationMomentumHeight
	w.ConfirmationsToFinality = aux.ConfirmationsToFinality
	return nil
}

// WrapTokenRequestList represents a paginated list of wrap requests.
//
// Fields:
//   - Count: Total number of wrap requests matching the query
//   - List: Slice of WrapTokenRequest entries for the current page
type WrapTokenRequestList struct {
	Count int                 `json:"count"`
	List  []*WrapTokenRequest `json:"list"`
}

// UnwrapTokenRequest represents an unwrap token request.
//
// When transferring tokens from an external chain to Zenon, an unwrap request
// is created after the external transaction is detected. Orchestrators verify
// and sign these requests to release tokens on Zenon.
//
// Fields:
//   - RegistrationMomentumHeight: Momentum height when request was registered
//   - NetworkClass: Category of source network
//   - ChainId: Chain identifier of source network
//   - TransactionHash: Transaction hash on the external chain
//   - LogIndex: Log index within the external transaction
//   - ToAddress: Recipient address on Zenon
//   - TokenAddress: Token contract address on the external chain
//   - TokenStandard: ZTS identifier of the token to receive
//   - Amount: Amount to unwrap (in base units, 8 decimals)
//   - Signature: TSS signature for the unwrap (added by orchestrators)
//   - Redeemed: Whether tokens have been claimed (1 = yes, 0 = no)
//   - Revoked: Whether request was revoked (1 = yes, 0 = no)
//   - RedeemableIn: Number of momentums until the request can be redeemed
type UnwrapTokenRequest struct {
	RegistrationMomentumHeight uint64                   `json:"registrationMomentumHeight"`
	NetworkClass               uint32                   `json:"networkClass"`
	ChainId                    uint32                   `json:"chainId"`
	TransactionHash            types.Hash               `json:"transactionHash"`
	LogIndex                   uint32                   `json:"logIndex"`
	ToAddress                  types.Address            `json:"toAddress"`
	TokenAddress               string                   `json:"tokenAddress"`
	TokenStandard              types.ZenonTokenStandard `json:"tokenStandard"`
	Amount                     *big.Int                 `json:"amount"`
	Signature                  string                   `json:"signature"`
	Redeemed                   uint32                   `json:"redeemed"`
	Revoked                    uint32                   `json:"revoked"`
	RedeemableIn               uint64                   `json:"redeemableIn"`
}

// unwrapTokenRequestJSON is used for JSON unmarshaling with string amounts
type unwrapTokenRequestJSON struct {
	RegistrationMomentumHeight uint64                   `json:"registrationMomentumHeight"`
	NetworkClass               uint32                   `json:"networkClass"`
	ChainId                    uint32                   `json:"chainId"`
	TransactionHash            types.Hash               `json:"transactionHash"`
	LogIndex                   uint32                   `json:"logIndex"`
	ToAddress                  types.Address            `json:"toAddress"`
	TokenAddress               string                   `json:"tokenAddress"`
	TokenStandard              types.ZenonTokenStandard `json:"tokenStandard"`
	Amount                     string                   `json:"amount"`
	Signature                  string                   `json:"signature"`
	Redeemed                   uint32                   `json:"redeemed"`
	Revoked                    uint32                   `json:"revoked"`
	RedeemableIn               uint64                   `json:"redeemableIn"`
}

func (u *UnwrapTokenRequest) UnmarshalJSON(data []byte) error {
	var aux unwrapTokenRequestJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	u.RegistrationMomentumHeight = aux.RegistrationMomentumHeight
	u.NetworkClass = aux.NetworkClass
	u.ChainId = aux.ChainId
	u.TransactionHash = aux.TransactionHash
	u.LogIndex = aux.LogIndex
	u.ToAddress = aux.ToAddress
	u.TokenAddress = aux.TokenAddress
	u.TokenStandard = aux.TokenStandard
	u.Amount = common.StringToBigInt(aux.Amount)
	u.Signature = aux.Signature
	u.Redeemed = aux.Redeemed
	u.Revoked = aux.Revoked
	u.RedeemableIn = aux.RedeemableIn
	return nil
}

// UnwrapTokenRequestList represents a paginated list of unwrap requests.
//
// Fields:
//   - Count: Total number of unwrap requests matching the query
//   - List: Slice of UnwrapTokenRequest entries for the current page
type UnwrapTokenRequestList struct {
	Count int                   `json:"count"`
	List  []*UnwrapTokenRequest `json:"list"`
}

// ZtsFeesInfo represents accumulated fees for a token.
//
// Bridge operations charge fees that accumulate in the bridge contract.
// This type tracks the total fees collected for a specific token.
//
// Fields:
//   - TokenStandard: ZTS identifier of the token
//   - AccumulatedFee: Total fees collected (in base units, 8 decimals)
type ZtsFeesInfo struct {
	TokenStandard  types.ZenonTokenStandard `json:"tokenStandard"`
	AccumulatedFee *big.Int                 `json:"accumulatedFee"`
}

// ztsFeesInfoJSON is used for JSON unmarshaling with string amounts
type ztsFeesInfoJSON struct {
	TokenStandard  types.ZenonTokenStandard `json:"tokenStandard"`
	AccumulatedFee string                   `json:"accumulatedFee"`
}

func (z *ZtsFeesInfo) UnmarshalJSON(data []byte) error {
	var aux ztsFeesInfoJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	z.TokenStandard = aux.TokenStandard
	z.AccumulatedFee = common.StringToBigInt(aux.AccumulatedFee)
	return nil
}
