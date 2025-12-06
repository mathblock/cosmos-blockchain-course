package marketplace

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	marketplacesimulation "skillchain/x/marketplace/simulation"
	"skillchain/x/marketplace/types"
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	marketplaceGenesis := types.GenesisState{
		Params: types.DefaultParams(),
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&marketplaceGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)
	const (
		opWeightMsgCreateProfile          = "op_weight_msg_marketplace"
		defaultWeightMsgCreateProfile int = 100
	)

	var weightMsgCreateProfile int
	simState.AppParams.GetOrGenerate(opWeightMsgCreateProfile, &weightMsgCreateProfile, nil,
		func(_ *rand.Rand) {
			weightMsgCreateProfile = defaultWeightMsgCreateProfile
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateProfile,
		marketplacesimulation.SimulateMsgCreateProfile(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgUpdateProfile          = "op_weight_msg_marketplace"
		defaultWeightMsgUpdateProfile int = 100
	)

	var weightMsgUpdateProfile int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateProfile, &weightMsgUpdateProfile, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateProfile = defaultWeightMsgUpdateProfile
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateProfile,
		marketplacesimulation.SimulateMsgUpdateProfile(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgCreateGig          = "op_weight_msg_marketplace"
		defaultWeightMsgCreateGig int = 100
	)

	var weightMsgCreateGig int
	simState.AppParams.GetOrGenerate(opWeightMsgCreateGig, &weightMsgCreateGig, nil,
		func(_ *rand.Rand) {
			weightMsgCreateGig = defaultWeightMsgCreateGig
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateGig,
		marketplacesimulation.SimulateMsgCreateGig(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgUpdateGigStatus          = "op_weight_msg_marketplace"
		defaultWeightMsgUpdateGigStatus int = 100
	)

	var weightMsgUpdateGigStatus int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateGigStatus, &weightMsgUpdateGigStatus, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateGigStatus = defaultWeightMsgUpdateGigStatus
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateGigStatus,
		marketplacesimulation.SimulateMsgUpdateGigStatus(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{}
}
