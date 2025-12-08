package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterInterfaces(registrar codectypes.InterfaceRegistry) {
	registrar.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateApplication{},
		&MsgUpdateApplication{},
		&MsgDeleteApplication{},
	)

	registrar.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateGigStatus{},
	)

	registrar.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateGig{},
	)

	registrar.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateProfile{},
	)

	registrar.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateProfile{},
	)

	registrar.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(registrar, &_Msg_serviceDesc)
}
