package types_test

import (
	"testing"

	"skillchain/x/marketplace/types"

	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc:     "valid genesis state",
			genState: &types.GenesisState{ProfileMap: []types.Profile{{Owner: "0"}, {Owner: "1"}}, GigList: []types.Gig{{Id: 0}, {Id: 1}}, GigCount: 2}, valid: true,
		}, {
			desc: "duplicated profile",
			genState: &types.GenesisState{
				ProfileMap: []types.Profile{
					{
						Owner: "0",
					},
					{
						Owner: "0",
					},
				},
				GigList: []types.Gig{{Id: 0}, {Id: 1}}, GigCount: 2,
			}, valid: false,
		}, {
			desc: "duplicated gig",
			genState: &types.GenesisState{
				GigList: []types.Gig{
					{
						Id: 0,
					},
					{
						Id: 0,
					},
				},
			},
			valid: false,
		}, {
			desc: "invalid gig count",
			genState: &types.GenesisState{
				GigList: []types.Gig{
					{
						Id: 1,
					},
				},
				GigCount: 0,
			},
			valid: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
