package keeper

import (
	"context"

	"skillchain/x/marketplace/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) ResolveDispute(ctx context.Context, msg *types.MsgResolveDispute) (*types.MsgResolveDisputeResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgResolveDisputeResponse{}, nil
}
