package keeper

import (
	"context"

	"fmt"
	"skillchain/x/marketplace/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) DeliverContract(goCtx context.Context, msg *types.MsgDeliverContract) (*types.MsgDeliverContractResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message
	ctx := sdk.UnwrapSDKContext(goCtx)

	contract, err := k.Contract.Get(ctx, msg.ContractId)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "contract %d not found", msg.ContractId)
	}

	if contract.Freelancer != msg.Creator {
		return nil, errorsmod.Wrap(types.ErrUnauthorized, "only freelancer can deliver")
	}

	if contract.Status != "active" {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"contract must be active to deliver (current: %s)",
			contract.Status,
		)
	}

	contract.Status = "delivered"
	err = k.Contract.Set(ctx, contract.Id, contract)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "failed to update contract status: %v", err)
	}

	// 5. Événement
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"contract_delivered",
			sdk.NewAttribute("contract_id", fmt.Sprintf("%d", msg.ContractId)),
			sdk.NewAttribute("freelancer", msg.Creator),
			sdk.NewAttribute("delivery_note", msg.DeliveryNote),
		),
	)

	return &types.MsgDeliverContractResponse{}, nil
}
