package embedded

import (
	"encoding/json"
	"math/big"

	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
)

// Token represents detailed information about a ZTS (Zenon Token Standard) token.
//
// ZTS tokens are native tokens on the Zenon Network. Any address can issue new
// tokens with customizable properties including supply limits, mintability,
// and burnability.
//
// Fields:
//   - Name: Human-readable name of the token (e.g., "Zenon")
//   - Symbol: Short ticker symbol (e.g., "ZNN")
//   - Domain: Associated domain for verification (optional)
//   - TotalSupply: Current circulating supply (in base units)
//   - Decimals: Number of decimal places (typically 8)
//   - Owner: Address that controls minting and token updates
//   - TokenStandard: Unique ZTS identifier for this token
//   - MaxSupply: Maximum possible supply (in base units)
//   - IsBurnable: Whether token holders can burn their tokens
//   - IsMintable: Whether the owner can mint additional tokens
//   - IsUtility: Whether this is a utility token
//
// Built-in Tokens:
//   - ZNN (types.ZnnTokenStandard): The native governance token
//   - QSR (types.QsrTokenStandard): The native utility token for plasma
//
// Example:
//
//	token, err := client.TokenApi.GetByZts(types.ZnnTokenStandard)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Token: %s (%s)\n", token.Name, token.Symbol)
//	fmt.Printf("Supply: %s / %s\n", token.TotalSupply, token.MaxSupply)
type Token struct {
	Name          string                   `json:"name"`
	Symbol        string                   `json:"symbol"`
	Domain        string                   `json:"domain"`
	TotalSupply   *big.Int                 `json:"totalSupply"`
	Decimals      uint8                    `json:"decimals"`
	Owner         types.Address            `json:"owner"`
	TokenStandard types.ZenonTokenStandard `json:"tokenStandard"`
	MaxSupply     *big.Int                 `json:"maxSupply"`
	IsBurnable    bool                     `json:"isBurnable"`
	IsMintable    bool                     `json:"isMintable"`
	IsUtility     bool                     `json:"isUtility"`
}

// tokenJSON is used for JSON unmarshaling with string amounts
type tokenJSON struct {
	Name          string                   `json:"name"`
	Symbol        string                   `json:"symbol"`
	Domain        string                   `json:"domain"`
	TotalSupply   string                   `json:"totalSupply"`
	Decimals      uint8                    `json:"decimals"`
	Owner         types.Address            `json:"owner"`
	TokenStandard types.ZenonTokenStandard `json:"tokenStandard"`
	MaxSupply     string                   `json:"maxSupply"`
	IsBurnable    bool                     `json:"isBurnable"`
	IsMintable    bool                     `json:"isMintable"`
	IsUtility     bool                     `json:"isUtility"`
}

func (t *Token) UnmarshalJSON(data []byte) error {
	var aux tokenJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	t.Name = aux.Name
	t.Symbol = aux.Symbol
	t.Domain = aux.Domain
	t.TotalSupply = common.StringToBigInt(aux.TotalSupply)
	t.Decimals = aux.Decimals
	t.Owner = aux.Owner
	t.TokenStandard = aux.TokenStandard
	t.MaxSupply = common.StringToBigInt(aux.MaxSupply)
	t.IsBurnable = aux.IsBurnable
	t.IsMintable = aux.IsMintable
	t.IsUtility = aux.IsUtility
	return nil
}

// TokenList represents a paginated list of tokens.
//
// This type is returned by methods that list multiple tokens, such as GetAll
// or GetByOwner.
//
// Fields:
//   - Count: Total number of tokens matching the query
//   - List: Slice of Token entries for the current page
type TokenList struct {
	Count int      `json:"count"`
	List  []*Token `json:"list"`
}
