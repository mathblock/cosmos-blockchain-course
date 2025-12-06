package types

import "fmt"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:     DefaultParams(),
		ProfileMap: []Profile{}, GigList: []Gig{}}
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

	return gs.Params.Validate()
}
