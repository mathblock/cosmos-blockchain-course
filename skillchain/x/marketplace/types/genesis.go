package types

import "fmt"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:     DefaultParams(),
		ProfileMap: []Profile{}, GigList: []Gig{}, ApplicationList: []Application{}, ContractList: []Contract{}, DisputeList: []Dispute{}, DisputeVoteMap: []DisputeVote{}}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	profileIndexMap := make(map[string]struct{})

	for _, elem := range gs.ProfileMap {
		index := fmt.Sprint(elem.Owner)
		if _, ok := profileIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for profile")
		}
		profileIndexMap[index] = struct{}{}
	}
	gigIdMap := make(map[uint64]bool)
	gigCount := gs.GetGigCount()
	for _, elem := range gs.GigList {
		if _, ok := gigIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for gig")
		}
		if elem.Id >= gigCount {
			return fmt.Errorf("gig id should be lower or equal than the last id")
		}
		gigIdMap[elem.Id] = true
	}
	applicationIdMap := make(map[uint64]bool)
	applicationCount := gs.GetApplicationCount()
	for _, elem := range gs.ApplicationList {
		if _, ok := applicationIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for application")
		}
		if elem.Id >= applicationCount {
			return fmt.Errorf("application id should be lower or equal than the last id")
		}
		applicationIdMap[elem.Id] = true
	}
	contractIdMap := make(map[uint64]bool)
	contractCount := gs.GetContractCount()
	for _, elem := range gs.ContractList {
		if _, ok := contractIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for contract")
		}
		if elem.Id >= contractCount {
			return fmt.Errorf("contract id should be lower or equal than the last id")
		}
		contractIdMap[elem.Id] = true
	}
	disputeIdMap := make(map[uint64]bool)
	disputeCount := gs.GetDisputeCount()
	for _, elem := range gs.DisputeList {
		if _, ok := disputeIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for dispute")
		}
		if elem.Id >= disputeCount {
			return fmt.Errorf("dispute id should be lower or equal than the last id")
		}
		disputeIdMap[elem.Id] = true
	}
	disputeVoteIndexMap := make(map[string]struct{})

	for _, elem := range gs.DisputeVoteMap {
		index := fmt.Sprint(elem.Arbiter)
		if _, ok := disputeVoteIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for disputeVote")
		}
		disputeVoteIndexMap[index] = struct{}{}
	}

	return gs.Params.Validate()
}
