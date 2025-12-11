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
				{
					RpcMethod: "ListApplication",
					Use:       "list-application",
					Short:     "List all application",
				},
				{
					RpcMethod:      "GetApplication",
					Use:            "get-application [id]",
					Short:          "Gets a application by id",
					Alias:          []string{"show-application"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod: "ListContract",
					Use:       "list-contract",
					Short:     "List all contract",
				},
				{
					RpcMethod:      "GetContract",
					Use:            "get-contract [id]",
					Short:          "Gets a contract by id",
					Alias:          []string{"show-contract"},
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
				{
					RpcMethod:      "CreateGig",
					Use:            "create-gig [title] [description] [price] [category] [delivery-days]",
					Short:          "Send a create-gig tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "title"}, {ProtoField: "description"}, {ProtoField: "price"}, {ProtoField: "category"}, {ProtoField: "delivery_days"}},
				},
				{
					RpcMethod:      "UpdateGigStatus",
					Use:            "update-gig-status [gig-id] [status]",
					Short:          "Send a update-gig-status tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "gig_id"}, {ProtoField: "status"}},
				},
				{
					RpcMethod:      "CreateApplication",
					Use:            "create-application [gig-id] [freelancer] [cover-letter] [proposed-price] [proposed-days] [status] [created-at]",
					Short:          "Create application",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "gig_id"}, {ProtoField: "freelancer"}, {ProtoField: "cover_letter"}, {ProtoField: "proposed_price"}, {ProtoField: "proposed_days"}, {ProtoField: "status"}, {ProtoField: "created_at"}},
				},
				{
					RpcMethod:      "UpdateApplication",
					Use:            "update-application [id] [gig-id] [freelancer] [cover-letter] [proposed-price] [proposed-days] [status] [created-at]",
					Short:          "Update application",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}, {ProtoField: "gig_id"}, {ProtoField: "freelancer"}, {ProtoField: "cover_letter"}, {ProtoField: "proposed_price"}, {ProtoField: "proposed_days"}, {ProtoField: "status"}, {ProtoField: "created_at"}},
				},
				{
					RpcMethod:      "DeleteApplication",
					Use:            "delete-application [id]",
					Short:          "Delete application",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "CreateContract",
					Use:            "create-contract [gig-id] [application-id] [client] [freelancer] [price] [delivery-deadline] [status] [created-at] [completed-at]",
					Short:          "Create contract",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "gig_id"}, {ProtoField: "application_id"}, {ProtoField: "client"}, {ProtoField: "freelancer"}, {ProtoField: "price"}, {ProtoField: "delivery_deadline"}, {ProtoField: "status"}, {ProtoField: "created_at"}, {ProtoField: "completed_at"}},
				},
				{
					RpcMethod:      "UpdateContract",
					Use:            "update-contract [id] [gig-id] [application-id] [client] [freelancer] [price] [delivery-deadline] [status] [created-at] [completed-at]",
					Short:          "Update contract",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}, {ProtoField: "gig_id"}, {ProtoField: "application_id"}, {ProtoField: "client"}, {ProtoField: "freelancer"}, {ProtoField: "price"}, {ProtoField: "delivery_deadline"}, {ProtoField: "status"}, {ProtoField: "created_at"}, {ProtoField: "completed_at"}},
				},
				{
					RpcMethod:      "DeleteContract",
					Use:            "delete-contract [id]",
					Short:          "Delete contract",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "ApplyToGig",
					Use:            "apply-to-gig [gig-id] [cover-letter] [proposed-price] [proposed-days]",
					Short:          "Send a apply-to-gig tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "gig_id"}, {ProtoField: "cover_letter"}, {ProtoField: "proposed_price"}, {ProtoField: "proposed_days"}},
				},
				{
					RpcMethod:      "WithdrawApplication",
					Use:            "withdraw-application [application-id]",
					Short:          "Send a withdraw-application tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "application_id"}},
				},
				{
					RpcMethod:      "AcceptApplication",
					Use:            "accept-application [application-id]",
					Short:          "Send a accept-application tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "application_id"}},
				},
				{
					RpcMethod:      "RejectApplication",
					Use:            "reject-application [application-id]",
					Short:          "Send a reject-application tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "application_id"}},
				},
				{
					RpcMethod:      "DeliverContract",
					Use:            "deliver-contract [contract-id] [delivery-note]",
					Short:          "Send a deliver-contract tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "contract_id"}, {ProtoField: "delivery_note"}},
				},
				{
					RpcMethod:      "CompleteContract",
					Use:            "complete-contract [contract-id]",
					Short:          "Send a complete-contract tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "contract_id"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
