package keeper

import (
	"context"
	"errors"
	"strings"

	"skillchain/x/marketplace/types"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CreateProfile(goCtx context.Context, msg *types.MsgCreateProfile) (*types.MsgCreateProfileResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := k.Profile.Get(ctx, msg.Creator)
	if err == nil {
		// Profile exists
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "profile for creator %s already exists", msg.Creator)
	} else if !errors.Is(err, collections.ErrNotFound) {
		// Some other error occurred
		return nil, errorsmod.Wrapf(err, "failed to check existing profile for creator %s", msg.Creator)
	}

	if msg.HourlyRate < 1 {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "hourly rate must be at least 1 skill")
    }

	if len(msg.Skills) < 1 {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "at least one skill is required")
    }

	profile := types.Profile{
        Owner:       msg.Creator,
        Name:        msg.Name,
        Bio:         msg.Bio,
        Skills:      msg.Skills,
        HourlyRate:  msg.HourlyRate,
        TotalJobs:   0,
        TotalEarned: 0,
        RatingSum:   0,
        RatingCount: 0,
    }

	err = k.Profile.Set(ctx, profile.Owner, profile)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "failed to create profile for creator %s", msg.Creator)
	}

	ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "profile_created",
            sdk.NewAttribute("owner", msg.Creator),
            sdk.NewAttribute("name", msg.Name),
			sdk.NewAttribute("bio", msg.Bio),
			sdk.NewAttribute("hourly_rate", string(msg.HourlyRate)),
			sdk.NewAttribute("skills", string(strings.Join(msg.Skills, ", "))),
        ),
    )

	return &types.MsgCreateProfileResponse{}, nil
}
