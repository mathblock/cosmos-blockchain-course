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
	for _, elem := range genState.GigList {
		if err := k.Gig.Set(ctx, elem.Id, elem); err != nil {
			return err
		}
	}

	if err := k.GigSeq.Set(ctx, genState.GigCount); err != nil {
		return err
	}
	for _, elem := range genState.ApplicationList {
		if err := k.Application.Set(ctx, elem.Id, elem); err != nil {
			return err
		}
	}

	if err := k.ApplicationSeq.Set(ctx, genState.ApplicationCount); err != nil {
		return err
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
	err = k.Gig.Walk(ctx, nil, func(key uint64, elem types.Gig) (bool, error) {
		genesis.GigList = append(genesis.GigList, elem)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	genesis.GigCount, err = k.GigSeq.Peek(ctx)
	if err != nil {
		return nil, err
	}
	err = k.Application.Walk(ctx, nil, func(key uint64, elem types.Application) (bool, error) {
		genesis.ApplicationList = append(genesis.ApplicationList, elem)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	genesis.ApplicationCount, err = k.ApplicationSeq.Peek(ctx)
	if err != nil {
		return nil, err
	}

	return genesis, nil
}
