package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	math "cosmossdk.io/math"

	"skillchain/x/marketplace/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) ResolveDispute(goCtx context.Context, msg *types.MsgResolveDispute) (*types.MsgResolveDisputeResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message
	ctx := sdk.UnwrapSDKContext(goCtx)
	dispute, errorDispute := k.Dispute.Get(ctx, msg.DisputeId)
	if errorDispute != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "dispute %d not found", msg.DisputeId)
	}

	if dispute.Status != "open" && dispute.Status != "voting" {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"dispute is not open for resolution (status: %s)",
			dispute.Status,
		)
	}

	contractId := dispute.ContractId

	contract, errorContract := k.Contract.Get(ctx, contractId)
    if errorContract != nil {
        return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "contract %d not found", contractId)
    }
    
    var winner string
    var winnerAddr sdk.AccAddress
    var errorWinner error
    
    if dispute.VotesClient > dispute.VotesFreelancer {
        winner = "client"
        dispute.Status = "resolved_client"
        dispute.Resolution = "Client wins by majority vote"
        winnerAddr, errorWinner = sdk.AccAddressFromBech32(contract.Client)
    } else if dispute.VotesFreelancer > dispute.VotesClient {
        winner = "freelancer"
        dispute.Status = "resolved_freelancer"
        dispute.Resolution = "Freelancer wins by majority vote"
        winnerAddr, errorWinner = sdk.AccAddressFromBech32(contract.Freelancer)
    } else {
        // If tie, favor the freelancer as per platform policy
        winner = "freelancer"
        dispute.Status = "resolved_freelancer"
        dispute.Resolution = "Tie resolved in favor of freelancer"
        winnerAddr, errorWinner = sdk.AccAddressFromBech32(contract.Freelancer)
    }
    
    if errorWinner != nil {
        return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid winner address")
    }
    
    escrowAmount := sdk.NewCoins(sdk.NewCoin("skill", math.NewIntFromUint64(contract.Price)))
    
    errSendCoin := k.bankKeeper.SendCoinsFromModuleToAccount(
        ctx,
        types.ModuleName,
        winnerAddr,
        escrowAmount,
    )
    if errSendCoin != nil {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to send coins to winner")
    }
    
    err := k.Dispute.Set(ctx, dispute.Id, dispute)
    if err != nil {
        return nil, errorsmod.Wrap(err, "failed to update dispute")
    }
    
    contract.Status = "resolved_" + winner
    contract.CompletedAt = ctx.BlockTime().Unix()
    err = k.Contract.Set(ctx, contract.Id, contract)
    if err != nil {
        return nil, errorsmod.Wrap(err, "failed to update contract")
    }
    
    gig, err := k.Gig.Get(ctx, contract.GigId)
    if err != nil {
        return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "gig %d not found", contract.GigId)
    }

    gig.Status = "closed"
    err = k.Gig.Set(ctx, gig.Id, gig)
    if err != nil {
        return nil, errorsmod.Wrap(err, "failed to update gig")
    }
    
    if winner == "freelancer" {
        profile, err := k.Profile.Get(ctx, contract.Freelancer)
        if err == nil {
            profile.TotalJobs++
            profile.TotalEarned += contract.Price
            k.Profile.Set(ctx, profile.Owner, profile)
        }
    }
    
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "dispute_resolved",
            sdk.NewAttribute("dispute_id", fmt.Sprintf("%d", dispute.Id)),
            sdk.NewAttribute("contract_id", fmt.Sprintf("%d", contract.Id)),
            sdk.NewAttribute("winner", winner),
            sdk.NewAttribute("amount", escrowAmount.String()),
        ),
    )

	return &types.MsgResolveDisputeResponse{}, nil
}
