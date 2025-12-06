package keeper_test

import (
	"testing"

	"skillchain/x/marketplace/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params:     types.DefaultParams(),
		ProfileMap: []types.Profile{{Owner: "0"}, {Owner: "1"}}, GigList: []types.Gig{{Id: 0}, {Id: 1}},
		GigCount: 2,
	}
	f := initFixture(t)
	err := f.keeper.InitGenesis(f.ctx, genesisState)
	require.NoError(t, err)
	got, err := f.keeper.ExportGenesis(f.ctx)
	require.NoError(t, err)
	require.NotNil(t, got)

	require.EqualExportedValues(t, genesisState.Params, got.Params)
	require.EqualExportedValues(t, genesisState.ProfileMap, got.ProfileMap)
	require.EqualExportedValues(t, genesisState.GigList, got.GigList)
	require.Equal(t, genesisState.GigCount, got.GigCount)

}
