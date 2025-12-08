package keeper

import (
	"context"
	"strings"

	"skillchain/x/marketplace/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) UpdateProfile(goCtx context.Context, msg *types.MsgUpdateProfile) (*types.MsgUpdateProfileResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := k.Profile.Get(ctx, msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "profile for creator %s not found", msg.Creator)
	}

	if msg.HourlyRate < 1 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "hourly rate must be at least 1 skill")
	}

	if len(msg.Skills) < 1 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "at least one skill is required")
	}

	profile := types.Profile{
		Owner:      msg.Creator,
		Name:       msg.Name,
		Bio:        msg.Bio,
		Skills:     msg.Skills,
		HourlyRate: msg.HourlyRate,
	}

	err = k.Profile.Set(ctx, profile.Owner, profile)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "failed to update profile for creator %s", msg.Creator)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"profile_updated",
			sdk.NewAttribute("owner", msg.Creator),
			sdk.NewAttribute("name", msg.Name),
			sdk.NewAttribute("bio", msg.Bio),
			sdk.NewAttribute("hourly_rate", string(msg.HourlyRate)),
			sdk.NewAttribute("skills", string(strings.Join(msg.Skills, ", "))),
		),
	)

	return &types.MsgUpdateProfileResponse{}, nil
}
