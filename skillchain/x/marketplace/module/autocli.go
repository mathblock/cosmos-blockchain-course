package marketplace

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"skillchain/x/marketplace/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: types.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the module",
				},
				{
					RpcMethod: "ListProfile",
					Use:       "list-profile",
					Short:     "List all profile",
				},
				{
					RpcMethod:      "GetProfile",
					Use:            "get-profile [id]",
					Short:          "Gets a profile",
					Alias:          []string{"show-profile"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "owner"}},
				},
				{
					RpcMethod: "ListGig",
					Use:       "list-gig",
					Short:     "List all gig",
				},
				{
					RpcMethod:      "GetGig",
					Use:            "get-gig [id]",
					Short:          "Gets a gig by id",
					Alias:          []string{"show-gig"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				// this line is used by ignite scaffolding # autocli/query
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              types.Msg_serviceDesc.ServiceName,
			EnhanceCustomCommand: true, // only required if you want to use the custom command
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
				{
					RpcMethod:      "CreateProfile",
					Use:            "create-profile [name] [bio] [skills] [hourly-rate]",
					Short:          "Send a create-profile tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "name"}, {ProtoField: "bio"}, {ProtoField: "skills"}, {ProtoField: "hourly_rate"}},
				},
				{
					RpcMethod:      "UpdateProfile",
					Use:            "update-profile [name] [bio] [skills] [hourly-rate]",
					Short:          "Send a update-profile tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "name"}, {ProtoField: "bio"}, {ProtoField: "skills"}, {ProtoField: "hourly_rate"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
