package types

import "fmt"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:     DefaultParams(),
		ProfileMap: []Profile{}}
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

	return gs.Params.Validate()
}
