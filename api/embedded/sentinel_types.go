package embedded

import (
	"github.com/zenon-network/go-zenon/common/types"
)

// SentinelInfo represents detailed information about a Sentinel.
//
// Sentinels are infrastructure nodes that support the Zenon Network by
// providing additional network services. They require a collateral deposit
// of ZNN and QSR and earn rewards for their participation.
//
// Fields:
//   - Owner: Address that owns and controls the Sentinel
//   - RegistrationTimestamp: Unix timestamp when the Sentinel was registered
//   - IsRevocable: Whether the Sentinel can be revoked (after cooldown period)
//   - RevokeCooldown: Remaining cooldown time before revocation completes
//   - Active: Whether the Sentinel is currently active and earning rewards
//
// Collateral Requirements:
//   - 5,000 ZNN
//   - 50,000 QSR
//
// Example:
//
//	sentinel, err := client.SentinelApi.GetByOwner(address)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if sentinel.Active {
//	    fmt.Printf("Sentinel owned by %s is active\n", sentinel.Owner)
//	}
type SentinelInfo struct {
	Owner                 types.Address `json:"owner"`
	RegistrationTimestamp int64         `json:"registrationTimestamp"`
	IsRevocable           bool          `json:"isRevocable"`
	RevokeCooldown        int64         `json:"revokeCooldown"`
	Active                bool          `json:"active"`
}

// SentinelInfoList represents a paginated list of Sentinels.
//
// This type is returned by methods that list multiple Sentinels, such as GetAllActive.
//
// Fields:
//   - Count: Total number of Sentinels matching the query
//   - List: Slice of SentinelInfo entries for the current page
type SentinelInfoList struct {
	Count int             `json:"count"`
	List  []*SentinelInfo `json:"list"`
}
