package keeper

import (
	"context"
	"fmt"

	"skillchain/x/marketplace/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) WithdrawApplication(goCtx context.Context, msg *types.MsgWithdrawApplication) (*types.MsgWithdrawApplicationResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message
	ctx := sdk.UnwrapSDKContext(goCtx)

	application, err := k.Application.Get(ctx, msg.ApplicationId)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "application with id %d not found: %v", msg.ApplicationId, err)
	}

	if application.Freelancer != msg.Creator {
		return nil, errorsmod.Wrap(types.ErrUnauthorized, "only the freelancer can withdraw their application")
	}

	if application.Status != "pending" {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"can only withdraw pending applications (current status: %s)",
			application.Status,
		)
	}

	application.Status = "withdrawn"
	err = k.Application.Set(ctx, application.Id, application)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "failed to update application status: %v", err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"application_withdrawn",
			sdk.NewAttribute("application_id", fmt.Sprintf("%d", application.Id)),
			sdk.NewAttribute("freelancer", application.Freelancer),
		),
	)

	return &types.MsgWithdrawApplicationResponse{}, nil
}
