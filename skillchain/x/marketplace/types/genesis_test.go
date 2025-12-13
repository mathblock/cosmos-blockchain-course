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
			genState: &types.GenesisState{ProfileMap: []types.Profile{{Owner: "0"}, {Owner: "1"}}, GigList: []types.Gig{{Id: 0}, {Id: 1}}, GigCount: 2, ApplicationList: []types.Application{{Id: 0}, {Id: 1}}, ApplicationCount: 2, ContractList: []types.Contract{{Id: 0}, {Id: 1}}, ContractCount: 2, DisputeList: []types.Dispute{{Id: 0}, {Id: 1}}, DisputeCount: 2, DisputeVoteMap: []types.DisputeVote{{Arbiter: "0"}, {Arbiter: "1"}}}, valid: true,
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
				ApplicationList: []types.Application{{Id: 0}, {Id: 1}}, ApplicationCount: 2, ContractList: []types.Contract{{Id: 0}, {Id: 1}}, ContractCount: 2, DisputeList: []types.Dispute{{Id: 0}, {Id: 1}}, DisputeCount: 2, DisputeVoteMap: []types.DisputeVote{{Arbiter: "0"}, {Arbiter: "1"}}}, valid: false,
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
				ApplicationList: []types.Application{{Id: 0}, {Id: 1}}, ApplicationCount: 2,
				ContractList: []types.Contract{{Id: 0}, {Id: 1}}, ContractCount: 2, DisputeList: []types.Dispute{{Id: 0}, {Id: 1}}, DisputeCount: 2, DisputeVoteMap: []types.DisputeVote{{Arbiter: "0"}, {Arbiter: "1"}}}, valid: false,
		}, {
			desc: "invalid gig count",
			genState: &types.GenesisState{
				GigList: []types.Gig{
					{
						Id: 1,
					},
				},
				GigCount:        0,
				ApplicationList: []types.Application{{Id: 0}, {Id: 1}}, ApplicationCount: 2,
				ContractList: []types.Contract{{Id: 0}, {Id: 1}}, ContractCount: 2, DisputeList: []types.Dispute{{Id: 0}, {Id: 1}}, DisputeCount: 2, DisputeVoteMap: []types.DisputeVote{{Arbiter: "0"}, {Arbiter: "1"}}}, valid: false,
		}, {
			desc: "duplicated application",
			genState: &types.GenesisState{
				ApplicationList: []types.Application{
					{
						Id: 0,
					},
					{
						Id: 0,
					},
				},
				ContractList: []types.Contract{{Id: 0}, {Id: 1}}, ContractCount: 2,
				DisputeList: []types.Dispute{{Id: 0}, {Id: 1}}, DisputeCount: 2, DisputeVoteMap: []types.DisputeVote{{Arbiter: "0"}, {Arbiter: "1"}}}, valid: false,
		}, {
			desc: "invalid application count",
			genState: &types.GenesisState{
				ApplicationList: []types.Application{
					{
						Id: 1,
					},
				},
				ApplicationCount: 0,
				ContractList:     []types.Contract{{Id: 0}, {Id: 1}}, ContractCount: 2,
				DisputeList: []types.Dispute{{Id: 0}, {Id: 1}}, DisputeCount: 2, DisputeVoteMap: []types.DisputeVote{{Arbiter: "0"}, {Arbiter: "1"}}}, valid: false,
		}, {
			desc: "duplicated contract",
			genState: &types.GenesisState{
				ContractList: []types.Contract{
					{
						Id: 0,
					},
					{
						Id: 0,
					},
				},
				DisputeList: []types.Dispute{{Id: 0}, {Id: 1}}, DisputeCount: 2,
				DisputeVoteMap: []types.DisputeVote{{Arbiter: "0"}, {Arbiter: "1"}}}, valid: false,
		}, {
			desc: "invalid contract count",
			genState: &types.GenesisState{
				ContractList: []types.Contract{
					{
						Id: 1,
					},
				},
				ContractCount: 0,
				DisputeList:   []types.Dispute{{Id: 0}, {Id: 1}}, DisputeCount: 2,
				DisputeVoteMap: []types.DisputeVote{{Arbiter: "0"}, {Arbiter: "1"}}}, valid: false,
		}, {
			desc: "duplicated dispute",
			genState: &types.GenesisState{
				DisputeList: []types.Dispute{
					{
						Id: 0,
					},
					{
						Id: 0,
					},
				},
				DisputeVoteMap: []types.DisputeVote{{Arbiter: "0"}, {Arbiter: "1"}}},
			valid: false,
		}, {
			desc: "invalid dispute count",
			genState: &types.GenesisState{
				DisputeList: []types.Dispute{
					{
						Id: 1,
					},
				},
				DisputeCount:   0,
				DisputeVoteMap: []types.DisputeVote{{Arbiter: "0"}, {Arbiter: "1"}}},
			valid: false,
		}, {
			desc: "duplicated disputeVote",
			genState: &types.GenesisState{
				DisputeVoteMap: []types.DisputeVote{
					{
						Arbiter: "0",
					},
					{
						Arbiter: "0",
					},
				},
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
