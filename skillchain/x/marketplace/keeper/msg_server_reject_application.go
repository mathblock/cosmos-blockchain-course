package keeper

import (
	"context"
	"fmt"

	"skillchain/x/marketplace/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) RejectApplication(goCtx context.Context, msg *types.MsgRejectApplication) (*types.MsgRejectApplicationResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message
	ctx := sdk.UnwrapSDKContext(goCtx)

	application, err := k.Application.Get(ctx, msg.ApplicationId)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "application with id %d not found: %v", msg.ApplicationId, err)
	}

	gig, err := k.Gig.Get(ctx, application.GigId)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "gig with id %d not found: %v", application.GigId, err)
	}

	if gig.Owner != msg.Creator {
		return nil, errorsmod.Wrap(types.ErrUnauthorized, "only the owner can reject applications for their gig")
	}

	if application.Status != "pending" {
        return nil, errorsmod.Wrapf(
            sdkerrors.ErrInvalidRequest,
            "can only reject pending applications (current status: %s)",
            application.Status,
        )
    }

	application.Status = "rejected"
	err = k.Application.Set(ctx, application.Id, application)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "failed to update application status: %v", err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"application_rejected",
			sdk.NewAttribute("application_id", fmt.Sprintf("%d", application.Id)),
			sdk.NewAttribute("owner", gig.Owner),
			sdk.NewAttribute("gig_id", fmt.Sprintf("%d", application.GigId)),
		),
	)

	return &types.MsgRejectApplicationResponse{}, nil
}
