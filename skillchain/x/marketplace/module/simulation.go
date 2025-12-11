package marketplace

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"skillchain/testutil/sample"
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
		Params:          types.DefaultParams(),
		ApplicationList: []types.Application{{Id: 0, Creator: sample.AccAddress()}, {Id: 1, Creator: sample.AccAddress()}}, ApplicationCount: 2,
		ContractList: []types.Contract{{Id: 0, Creator: sample.AccAddress()}, {Id: 1, Creator: sample.AccAddress()}}, ContractCount: 2,
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
	const (
		opWeightMsgCreateApplication          = "op_weight_msg_marketplace"
		defaultWeightMsgCreateApplication int = 100
	)

	var weightMsgCreateApplication int
	simState.AppParams.GetOrGenerate(opWeightMsgCreateApplication, &weightMsgCreateApplication, nil,
		func(_ *rand.Rand) {
			weightMsgCreateApplication = defaultWeightMsgCreateApplication
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateApplication,
		marketplacesimulation.SimulateMsgCreateApplication(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgUpdateApplication          = "op_weight_msg_marketplace"
		defaultWeightMsgUpdateApplication int = 100
	)

	var weightMsgUpdateApplication int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateApplication, &weightMsgUpdateApplication, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateApplication = defaultWeightMsgUpdateApplication
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateApplication,
		marketplacesimulation.SimulateMsgUpdateApplication(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgDeleteApplication          = "op_weight_msg_marketplace"
		defaultWeightMsgDeleteApplication int = 100
	)

	var weightMsgDeleteApplication int
	simState.AppParams.GetOrGenerate(opWeightMsgDeleteApplication, &weightMsgDeleteApplication, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteApplication = defaultWeightMsgDeleteApplication
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteApplication,
		marketplacesimulation.SimulateMsgDeleteApplication(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgCreateContract          = "op_weight_msg_marketplace"
		defaultWeightMsgCreateContract int = 100
	)

	var weightMsgCreateContract int
	simState.AppParams.GetOrGenerate(opWeightMsgCreateContract, &weightMsgCreateContract, nil,
		func(_ *rand.Rand) {
			weightMsgCreateContract = defaultWeightMsgCreateContract
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateContract,
		marketplacesimulation.SimulateMsgCreateContract(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgUpdateContract          = "op_weight_msg_marketplace"
		defaultWeightMsgUpdateContract int = 100
	)

	var weightMsgUpdateContract int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateContract, &weightMsgUpdateContract, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateContract = defaultWeightMsgUpdateContract
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateContract,
		marketplacesimulation.SimulateMsgUpdateContract(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgDeleteContract          = "op_weight_msg_marketplace"
		defaultWeightMsgDeleteContract int = 100
	)

	var weightMsgDeleteContract int
	simState.AppParams.GetOrGenerate(opWeightMsgDeleteContract, &weightMsgDeleteContract, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteContract = defaultWeightMsgDeleteContract
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteContract,
		marketplacesimulation.SimulateMsgDeleteContract(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgApplyToGig          = "op_weight_msg_marketplace"
		defaultWeightMsgApplyToGig int = 100
	)

	var weightMsgApplyToGig int
	simState.AppParams.GetOrGenerate(opWeightMsgApplyToGig, &weightMsgApplyToGig, nil,
		func(_ *rand.Rand) {
			weightMsgApplyToGig = defaultWeightMsgApplyToGig
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgApplyToGig,
		marketplacesimulation.SimulateMsgApplyToGig(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgWithdrawApplication          = "op_weight_msg_marketplace"
		defaultWeightMsgWithdrawApplication int = 100
	)

	var weightMsgWithdrawApplication int
	simState.AppParams.GetOrGenerate(opWeightMsgWithdrawApplication, &weightMsgWithdrawApplication, nil,
		func(_ *rand.Rand) {
			weightMsgWithdrawApplication = defaultWeightMsgWithdrawApplication
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgWithdrawApplication,
		marketplacesimulation.SimulateMsgWithdrawApplication(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{}
}
