package keeper

import (
	"context"

	"skillchain/x/marketplace/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var validStatusTransitions = map[string][]string{
    "open":        {"cancelled", "in_progress"},
    "in_progress": {"completed", "disputed"},
    "disputed":    {"completed", "cancelled"},
}

func (k msgServer) UpdateGigStatus(goCtx context.Context, msg *types.MsgUpdateGigStatus) (*types.MsgUpdateGigStatusResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message
	ctx := sdk.UnwrapSDKContext(goCtx)

	gig, err := k.Gig.Get(ctx, msg.GigId)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "gig with ID %d not found", msg.GigId)
	}

	if gig.Owner != msg.Creator {
        return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only gig owner can update status")
    }

	allowedTransitions, exists := validStatusTransitions[gig.Status]
    if !exists {
        return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "gig status %s cannot be changed", gig.Status)
    }
	isValidTransition := false
    for _, allowed := range allowedTransitions {
        if allowed == msg.Status {
            isValidTransition = true
            break
        }
    }
    if !isValidTransition {
        return nil, errorsmod.Wrapf(
            sdkerrors.ErrInvalidRequest,
            "cannot transition from %s to %s",
            gig.Status,
            msg.Status,
        )
    }

	gig.Status = msg.Status

	err = k.Gig.Set(ctx, gig.Id, gig)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "failed to update gig status for gig ID %d", msg.GigId)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"gig_status_updated",
			sdk.NewAttribute("gig_id", string(msg.GigId)),
			sdk.NewAttribute("new_status", msg.Status),
		),
	)

	return &types.MsgUpdateGigStatusResponse{}, nil
}
