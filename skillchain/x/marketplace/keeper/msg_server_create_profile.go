package keeper

import (
	"context"

	"skillchain/x/marketplace/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) CreateProfile(ctx context.Context, msg *types.MsgCreateProfile) (*types.MsgCreateProfileResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgCreateProfileResponse{}, nil
}
