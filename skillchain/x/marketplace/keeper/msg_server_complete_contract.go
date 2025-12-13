package keeper

import (
	"context"
	"fmt"
	"skillchain/x/marketplace/types"

	errorsmod "cosmossdk.io/errors"
	math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CompleteContract(goCtx context.Context, msg *types.MsgCompleteContract) (*types.MsgCompleteContractResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message
	ctx := sdk.UnwrapSDKContext(goCtx)

	contract, err := k.Contract.Get(ctx, msg.ContractId)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "contract %d not found", msg.ContractId)
	}

	if contract.Client != msg.Creator {
		return nil, errorsmod.Wrap(types.ErrUnauthorized, "only client can complete the contract")
	}

	if contract.Status != "delivered" {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"contract must be delivered to complete (current: %s)",
			contract.Status,
		)
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "failed to get params")
	}
	totalAmount := math.NewIntFromUint64(contract.Price)

	// Platform Fee = price * feePercent / 100
	platformFee := totalAmount.Mul(math.NewIntFromUint64(params.PlatformFeePercent)).Quo(math.NewInt(100))
	freelancerAmount := totalAmount.Sub(platformFee)

	freelancerAddr, err := sdk.AccAddressFromBech32(contract.Freelancer)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid freelancer address")
	}

	freelancerCoins := sdk.NewCoins(sdk.NewCoin("skill", freelancerAmount))
	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		freelancerAddr,
		freelancerCoins,
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to release funds to freelancer")
	}

	contract.Status = "completed"
	contract.CompletedAt = ctx.BlockTime().Unix()
	err = k.Contract.Set(ctx, contract.Id, contract)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "failed to update contract status: %v", err)
	}

	gig, err := k.Gig.Get(ctx, contract.GigId)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "gig %d not found", contract.GigId)
	}
	gig.Status = "completed"
	err = k.Gig.Set(ctx, gig.Id, gig)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "failed to update gig status: %v", err)
	}

	profile, err := k.Profile.Get(ctx, contract.Freelancer)
	if err == nil {
		profile.TotalJobs++
		profile.TotalEarned += freelancerAmount.Uint64()
		err = k.Profile.Set(ctx, contract.Freelancer, profile)
		if err != nil {
			return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "failed to update profile: %v", err)
		}
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"contract_completed",
			sdk.NewAttribute("contract_id", fmt.Sprintf("%d", msg.ContractId)),
			sdk.NewAttribute("client", contract.Client),
			sdk.NewAttribute("freelancer", contract.Freelancer),
		),
		sdk.NewEvent(
			"payment_released",
			sdk.NewAttribute("contract_id", fmt.Sprintf("%d", msg.ContractId)),
			sdk.NewAttribute("freelancer", contract.Freelancer),
			sdk.NewAttribute("amount", freelancerCoins.String()),
			sdk.NewAttribute("platform_fee", platformFee.String()),
		),
	})

	return &types.MsgCompleteContractResponse{}, nil
}
