package keeper

import (
	"context"

	"skillchain/x/marketplace/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) ApplyToGig(goCtx context.Context, msg *types.MsgApplyToGig) (*types.MsgApplyToGigResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message
	ctx := sdk.UnwrapSDKContext(goCtx)

	gig, err := k.Gig.Get(ctx, msg.GigId)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrGigNotFound, "gig %d not found", msg.GigId)
	}

	if gig.Status != "open" {
        return nil, errorsmod.Wrapf(
            sdkerrors.ErrInvalidRequest, 
            "gig %d is not open for applications (status: %s)", 
            msg.GigId, 
            gig.Status,
        )
    }

	_, err = k.Profile.Get(ctx, msg.Creator)
    if err != nil {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "freelancer must have a profile to apply")
    }

	if gig.Owner == msg.Creator {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "cannot apply to your own gig")
    }
    
    applicationFound := false
    k.Application.Walk(ctx, nil, func(_ uint64, app types.Application) (stop bool, err error) {
        if app.GigId == msg.GigId && app.Freelancer == msg.Creator && app.Status == "pending" {
            applicationFound = true
            return true, nil
        }
        return false, nil
    })
    if applicationFound {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "you already have a pending application for this gig")
    }

    params, err := k.Params.Get(ctx)
    if err != nil {
        return nil, errorsmod.Wrap(err, "failed to get marketplace parameters")
    }

    if math.NewInt(int64(msg.ProposedPrice)).LT(params.MinGigPrice) {
        return nil, errorsmod.Wrap(types.ErrInvalidPrice, "proposed price is below minimum")
    }

	id, err := k.ApplicationSeq.Next(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get next application id")
	}

	application := types.Application{
		Id:         id,
		GigId:      msg.GigId,
		Freelancer: msg.Creator,
		ProposedPrice: msg.ProposedPrice,
		CoverLetter: msg.CoverLetter,
		Status:     "pending",
	}

	err = k.Application.Set(ctx, application.Id, application)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to save application")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"application_submitted",
			sdk.NewAttribute("application_id", string(application.Id)),
			sdk.NewAttribute("gig_id", string(application.GigId)),
			sdk.NewAttribute("freelancer", application.Freelancer),
			sdk.NewAttribute("proposed_price", string(application.ProposedPrice)),
			sdk.NewAttribute("status", application.Status),
		),
	)

    return &types.MsgApplyToGigResponse{
		ApplicationId: application.Id,
	}, nil
}
