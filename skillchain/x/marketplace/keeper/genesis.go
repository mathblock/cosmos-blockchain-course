package keeper

import (
	"context"

	"skillchain/x/marketplace/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
	for _, elem := range genState.ProfileMap {
		if err := k.Profile.Set(ctx, elem.Owner, elem); err != nil {
			return err
		}
	}

	return k.Params.Set(ctx, genState.Params)
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	var err error

	genesis := types.DefaultGenesis()
	genesis.Params, err = k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	if err := k.Profile.Walk(ctx, nil, func(_ string, val types.Profile) (stop bool, err error) {
		genesis.ProfileMap = append(genesis.ProfileMap, val)
		return false, nil
	}); err != nil {
		return nil, err
	}

	return genesis, nil
}
