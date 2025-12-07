package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"

	"skillchain/x/marketplace/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) CreateGig(goCtx context.Context, msg *types.MsgCreateGig) (*types.MsgCreateGigResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message
	ctx := sdk.UnwrapSDKContext(goCtx)

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get params")
	}

	priceInt := sdkmath.NewIntFromUint64(msg.Price)
	if priceInt.LT(params.MinGigPrice) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "gig price must be at least %d skill", params.MinGigPrice)
	}

	if msg.DeliveryDays < 1 || msg.DeliveryDays > 365 {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "delivery days must be between 1 and 365")
    }

	if len(msg.Title) < 10 || len(msg.Title) > 100 {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "title must be between 10 and 100 characters")
    }

	if len(msg.Description) < 50 || len(msg.Description) > 1000 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "description must be between 50 and 1000 characters")
	}

	id, err := k.GigSeq.Next(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get next gig id")
	}

	gig := types.Gig{
		Id:           id,
        Title:        msg.Title,
        Description:  msg.Description,
        Owner:        msg.Creator,
        Price:        msg.Price,
        Category:     msg.Category,
        DeliveryDays: msg.DeliveryDays,
        Status:       "open",
        CreatedAt:    ctx.BlockTime().Unix(),
    }

	err = k.Gig.Set(ctx, gig.Id, gig)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "failed to create gig for creator %s", msg.Creator)
	}

	ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "gig_created",
            sdk.NewAttribute("owner", msg.Creator),
            sdk.NewAttribute("title", msg.Title),
			sdk.NewAttribute("description", msg.Description),
			sdk.NewAttribute("price", string(msg.Price)),
			sdk.NewAttribute("category", msg.Category),
			sdk.NewAttribute("status", gig.Status),
			sdk.NewAttribute("delivery_days", string(msg.DeliveryDays)),
        ),
    )

	return &types.MsgCreateGigResponse{
		Id: gig.Id,
	}, nil
}
