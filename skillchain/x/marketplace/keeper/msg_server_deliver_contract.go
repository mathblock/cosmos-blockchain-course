package keeper

import (
	"context"

	"skillchain/x/marketplace/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) DeliverContract(ctx context.Context, msg *types.MsgDeliverContract) (*types.MsgDeliverContractResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgDeliverContractResponse{}, nil
}
