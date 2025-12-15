package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"skillchain/x/marketplace/types"
)

func (k Keeper) ProcessExpiredDisputes(ctx sdk.Context) error {
    currentTime := ctx.BlockTime().Unix()
   	err := k.Dispute.Walk(ctx, nil, func(key uint64, dispute types.Dispute) (stop bool, err error) {
		if dispute.Status != "open" && dispute.Status != "voting" {
			return false, nil
		}
        
		if dispute.Deadline > currentTime { 
			return false, nil
		}

		params, err := k.Params.Get(ctx)
		if err != nil {
			return true, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "failed to get params")
		}

		totalVotes := dispute.VotesClient + dispute.VotesFreelancer
		if totalVotes >= params.MinArbitersRequired {
			err := k.resolveDisputeInternal(ctx, dispute.Id)
			if err != nil {
				return true, errorsmod.Wrapf(err, "failed to resolve expired dispute %d", dispute.Id)
			}
		} else {
			if dispute.VotesFreelancer == 0 {
				dispute.VotesFreelancer = 1
			}
			dispute.Resolution = fmt.Sprintf(
				"Expired with insufficient votes (%d/%d required). Defaulting to freelancer.",
				totalVotes,
				params.MinArbitersRequired,
			)
			err := k.Dispute.Set(ctx, dispute.Id, dispute)
			if err != nil {
				return true, errorsmod.Wrapf(err, "failed to update dispute %d", dispute.Id)
			}

			err = k.resolveDisputeInternal(ctx, dispute.Id)
			if err != nil {
				return true, errorsmod.Wrapf(err, "failed to resolve expired dispute %d", dispute.Id)
			}
			
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"dispute_expired",
					sdk.NewAttribute("dispute_id", fmt.Sprintf("%d", dispute.Id)),
					sdk.NewAttribute("total_votes", fmt.Sprintf("%d", totalVotes)),
					sdk.NewAttribute("required_votes", fmt.Sprintf("%d", params.MinArbitersRequired)),
				),
			)
		} 
		return false, nil
	})
	
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "error processing expired disputes")
    }

    return nil
}