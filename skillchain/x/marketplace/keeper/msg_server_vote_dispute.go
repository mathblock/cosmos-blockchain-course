package keeper

import (
	"context"
	"fmt"

	math "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"skillchain/x/marketplace/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) VoteDispute(goCtx context.Context, msg *types.MsgVoteDispute) (*types.MsgVoteDisputeResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message
	ctx := sdk.UnwrapSDKContext(goCtx)
    
    dispute, err := k.Dispute.Get(ctx, msg.DisputeId)
    if err != nil {
        return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "dispute %d not found", msg.DisputeId)
    }
    
    if dispute.Status != "open" && dispute.Status != "voting" {
        return nil, errorsmod.Wrapf(
            sdkerrors.ErrInvalidRequest,
            "dispute is not open for voting (status: %s)",
            dispute.Status,
        )
    }
    
    if ctx.BlockTime().Unix() > dispute.Deadline {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "voting deadline has passed")
    }
    
    // 4. Vérifier que le voter n'est pas partie au contrat
    contract, err := k.Contract.Get(ctx, dispute.ContractId)
    if err != nil {
        return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "contract %d not found", dispute.ContractId)
    }
    if contract.Client == msg.Creator || contract.Freelancer == msg.Creator {
        return nil, errorsmod.Wrap(types.ErrUnauthorized, "parties cannot vote on their own dispute")
    }
    
    params, err := k.Params.Get(ctx)
    if err != nil {
        return nil, errorsmod.Wrap(err, "failed to get params")
    }
    voterAddr, _ := sdk.AccAddressFromBech32(msg.Creator)
    balance := k.bankKeeper.GetBalance(ctx, voterAddr, "stake")
    
    if balance.Amount.LT(math.NewIntFromUint64(params.ArbiterStakeRequired)) {
        return nil, errorsmod.Wrapf(
            types.ErrInsufficientFunds,
            "arbiter must have at least %d stake (has %s)",
            params.ArbiterStakeRequired,
            balance.String(),
        )
    }
    
    k.DisputeVote.Walk(ctx, nil, func(key string, disputeVote types.DisputeVote) (stop bool, err error) {
		if disputeVote.DisputeId == msg.DisputeId && disputeVote.Arbiter == msg.Creator {
			return true, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "already voted on this dispute")
		}
		return false, nil
	})
    
    
    if msg.Vote != "client" && msg.Vote != "freelancer" {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "vote must be 'client' or 'freelancer'")
    }
    
    vote := types.DisputeVote{
        Arbiter:   msg.Creator,
        DisputeId: msg.DisputeId,
        Vote:      msg.Vote,
        VotedAt:   ctx.BlockTime().Unix(),
    }

    err = k.DisputeVote.Set(ctx, vote.Arbiter, vote)
    if err != nil {
        return nil, errorsmod.Wrap(err, "failed to record dispute vote")
    }
    
    if msg.Vote == "client" {
        dispute.VotesClient++
    } else {
        dispute.VotesFreelancer++
    }

	winner := "freelancer"
    if dispute.VotesClient > dispute.VotesFreelancer {
        winner = "client"
    }
    
    if dispute.Status == "open" {
        dispute.Status = "voting"
    }
    
    err = k.Dispute.Set(ctx, dispute.Id, dispute)
    if err != nil {
        return nil, errorsmod.Wrap(err, "failed to update dispute")
    }

    totalVotes := dispute.VotesClient + dispute.VotesFreelancer
    if totalVotes >= params.MinArbitersRequired {
		msgResolveDispute := &types.MsgResolveDispute{
			Creator:   "AutomatedResolver",
			DisputeId: dispute.Id,
			Winner: winner,
		}
        k.ResolveDispute(ctx, msgResolveDispute)
    }
    
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "dispute_vote_cast",
            sdk.NewAttribute("dispute_id", fmt.Sprintf("%d", msg.DisputeId)),
            sdk.NewAttribute("arbiter", msg.Creator),
            sdk.NewAttribute("vote", msg.Vote),
            sdk.NewAttribute("total_votes", fmt.Sprintf("%d", totalVotes)),
        ),
    )

	return &types.MsgVoteDisputeResponse{}, nil
}
