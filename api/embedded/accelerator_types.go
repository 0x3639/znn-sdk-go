package embedded

import (
	"encoding/json"
	"math/big"

	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
)

// VoteBreakdown represents voting results for a project or phase.
//
// Pillars vote on Accelerator-Z projects and phases. This type contains
// the vote tally for a specific project or phase.
//
// Fields:
//   - Id: Hash of the project or phase being voted on
//   - Total: Total number of votes cast
//   - Yes: Number of approval votes
//   - No: Number of rejection votes
type VoteBreakdown struct {
	Id    types.Hash `json:"id"`
	Total uint32     `json:"total"`
	Yes   uint32     `json:"yes"`
	No    uint32     `json:"no"`
}

// PhaseInfo represents the phase details within a Phase.
//
// Accelerator-Z projects are divided into phases, each with its own funding
// requirements and approval process. This type contains the core information
// about a project phase.
//
// Fields:
//   - Id: Unique identifier for this phase
//   - ProjectID: Parent project's identifier
//   - Name: Human-readable name of the phase
//   - Description: Detailed description of phase deliverables
//   - Url: Link to additional information or documentation
//   - ZnnFundsNeeded: ZNN funding requested (in base units, 8 decimals)
//   - QsrFundsNeeded: QSR funding requested (in base units, 8 decimals)
//   - CreationTimestamp: Unix timestamp when phase was created
//   - AcceptedTimestamp: Unix timestamp when phase was accepted (0 if pending)
//   - Status: Current phase status
type PhaseInfo struct {
	Id                types.Hash `json:"id"`
	ProjectID         types.Hash `json:"projectID"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	Url               string     `json:"url"`
	ZnnFundsNeeded    *big.Int   `json:"znnFundsNeeded"`
	QsrFundsNeeded    *big.Int   `json:"qsrFundsNeeded"`
	CreationTimestamp int64      `json:"creationTimestamp"`
	AcceptedTimestamp int64      `json:"acceptedTimestamp"`
	Status            uint8      `json:"status"`
}

// phaseInfoJSON is used for JSON unmarshaling with string amounts
type phaseInfoJSON struct {
	Id                types.Hash `json:"id"`
	ProjectID         types.Hash `json:"projectID"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	Url               string     `json:"url"`
	ZnnFundsNeeded    string     `json:"znnFundsNeeded"`
	QsrFundsNeeded    string     `json:"qsrFundsNeeded"`
	CreationTimestamp int64      `json:"creationTimestamp"`
	AcceptedTimestamp int64      `json:"acceptedTimestamp"`
	Status            uint8      `json:"status"`
}

func (p *PhaseInfo) UnmarshalJSON(data []byte) error {
	var aux phaseInfoJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	p.Id = aux.Id
	p.ProjectID = aux.ProjectID
	p.Name = aux.Name
	p.Description = aux.Description
	p.Url = aux.Url
	p.ZnnFundsNeeded = common.StringToBigInt(aux.ZnnFundsNeeded)
	p.QsrFundsNeeded = common.StringToBigInt(aux.QsrFundsNeeded)
	p.CreationTimestamp = aux.CreationTimestamp
	p.AcceptedTimestamp = aux.AcceptedTimestamp
	p.Status = aux.Status
	return nil
}

// Phase represents a project phase with its voting info.
//
// This type combines the phase details with the current vote breakdown.
//
// Fields:
//   - Phase: Core phase information
//   - Votes: Current voting results for this phase
type Phase struct {
	Phase *PhaseInfo     `json:"phase"`
	Votes *VoteBreakdown `json:"votes"`
}

// Project represents an Accelerator-Z project.
//
// Accelerator-Z is Zenon's ecosystem fund that supports community development
// projects. Projects request funding in ZNN and QSR, and Pillars vote to
// approve or reject them.
//
// Fields:
//   - Id: Unique identifier for this project
//   - Owner: Address that submitted and manages the project
//   - Name: Human-readable name of the project
//   - Description: Detailed description of project goals and deliverables
//   - Url: Link to additional information or documentation
//   - ZnnFundsNeeded: Total ZNN funding requested (in base units, 8 decimals)
//   - QsrFundsNeeded: Total QSR funding requested (in base units, 8 decimals)
//   - CreationTimestamp: Unix timestamp when project was submitted
//   - LastUpdateTimestamp: Unix timestamp of last modification
//   - Status: Current project status
//   - PhaseIds: List of phase identifiers for this project
//   - Votes: Current voting results for the project
//   - Phases: Detailed phase information with voting data
//
// Example:
//
//	projects, err := client.AcceleratorApi.GetAll(0, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, project := range projects.List {
//	    fmt.Printf("Project: %s, Votes: %d/%d\n",
//	        project.Name, project.Votes.Yes, project.Votes.Total)
//	}
type Project struct {
	Id                  types.Hash     `json:"id"`
	Owner               types.Address  `json:"owner"`
	Name                string         `json:"name"`
	Description         string         `json:"description"`
	Url                 string         `json:"url"`
	ZnnFundsNeeded      *big.Int       `json:"znnFundsNeeded"`
	QsrFundsNeeded      *big.Int       `json:"qsrFundsNeeded"`
	CreationTimestamp   int64          `json:"creationTimestamp"`
	LastUpdateTimestamp int64          `json:"lastUpdateTimestamp"`
	Status              uint8          `json:"status"`
	PhaseIds            []types.Hash   `json:"phaseIds"`
	Votes               *VoteBreakdown `json:"votes"`
	Phases              []*Phase       `json:"phases"`
}

// projectJSON is used for JSON unmarshaling with string amounts
type projectJSON struct {
	Id                  types.Hash     `json:"id"`
	Owner               types.Address  `json:"owner"`
	Name                string         `json:"name"`
	Description         string         `json:"description"`
	Url                 string         `json:"url"`
	ZnnFundsNeeded      string         `json:"znnFundsNeeded"`
	QsrFundsNeeded      string         `json:"qsrFundsNeeded"`
	CreationTimestamp   int64          `json:"creationTimestamp"`
	LastUpdateTimestamp int64          `json:"lastUpdateTimestamp"`
	Status              uint8          `json:"status"`
	PhaseIds            []types.Hash   `json:"phaseIds"`
	Votes               *VoteBreakdown `json:"votes"`
	Phases              []*Phase       `json:"phases"`
}

func (p *Project) UnmarshalJSON(data []byte) error {
	var aux projectJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	p.Id = aux.Id
	p.Owner = aux.Owner
	p.Name = aux.Name
	p.Description = aux.Description
	p.Url = aux.Url
	p.ZnnFundsNeeded = common.StringToBigInt(aux.ZnnFundsNeeded)
	p.QsrFundsNeeded = common.StringToBigInt(aux.QsrFundsNeeded)
	p.CreationTimestamp = aux.CreationTimestamp
	p.LastUpdateTimestamp = aux.LastUpdateTimestamp
	p.Status = aux.Status
	p.PhaseIds = aux.PhaseIds
	p.Votes = aux.Votes
	// Correctly copy phases from aux (this fixes the go-zenon bug)
	p.Phases = aux.Phases
	return nil
}

// ProjectList represents a paginated list of projects.
//
// This type is returned by methods that list Accelerator-Z projects, such as
// GetAll.
//
// Fields:
//   - Count: Total number of projects matching the query
//   - List: Slice of Project entries for the current page
type ProjectList struct {
	Count int        `json:"count"`
	List  []*Project `json:"list"`
}
