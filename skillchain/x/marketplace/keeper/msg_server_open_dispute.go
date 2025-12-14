package keeper

import (
	"context"
	"fmt"

	"skillchain/x/marketplace/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) OpenDispute(goCtx context.Context, msg *types.MsgOpenDispute) (*types.MsgOpenDisputeResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message
	ctx := sdk.UnwrapSDKContext(goCtx)
    
    contract, err := k.Contract.Get(ctx, msg.ContractId)
    if err != nil {
        return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "contract %d not found", msg.ContractId)
    }
    
    isClient := contract.Client == msg.Creator
    isFreelancer := contract.Freelancer == msg.Creator
    
    if !isClient && !isFreelancer {
        return nil, errorsmod.Wrap(types.ErrUnauthorized, "only client or freelancer can open dispute")
    }
    
    if contract.Status != "active" && contract.Status != "delivered" {
        return nil, errorsmod.Wrapf(
            sdkerrors.ErrInvalidRequest,
            "cannot dispute contract with status %s",
            contract.Status,
        )
    }
    
    err = k.Dispute.Walk(ctx, nil, func(_ uint64, dispute types.Dispute) (bool, error) {
        if dispute.ContractId == msg.ContractId && dispute.Status == "open" {
            err = errorsmod.Wrapf(
                sdkerrors.ErrInvalidRequest,
                "there is already an open dispute for contract %d",
                msg.ContractId,
            )
            return true, nil 
        }
        return false, nil
    })
    if err != nil {
        return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "contract %d not found", msg.ContractId)
    }
    
    params, err := k.Params.Get(ctx)
    if err != nil {
        return nil, errorsmod.Wrap(err, "failed to get marketplace params")
    }
    deadline := ctx.BlockTime().Unix() + int64(params.DisputeDuration)
    
    dispute := types.Dispute{
        ContractId:         msg.ContractId,
        Initiator:          msg.Creator,
        Reason:             msg.Reason,
        Status:             "open",
        VotesClient:        0,
        VotesFreelancer:    0,
        Resolution:         "",
        CreatedAt:          ctx.BlockTime().Unix(),
        Deadline:           deadline,
    }
    
    if isClient {
        dispute.ClientEvidence = msg.Evidence
    } else {
        dispute.FreelancerEvidence = msg.Evidence
    }
    
    disputeId, err := k.DisputeSeq.Next(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get next dispute id")
	}
	err = k.Dispute.Set(ctx, disputeId, dispute)
    if err != nil {
        return nil, errorsmod.Wrap(err, "failed to set dispute")
    }
    
    contract.Status = "disputed"
    err = k.Contract.Set(ctx, contract.Id, contract)
    if err != nil {
        return nil, errorsmod.Wrap(err, "failed to update contract status")
    }
    
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "dispute_opened",
            sdk.NewAttribute("dispute_id", fmt.Sprintf("%d", disputeId)),
            sdk.NewAttribute("contract_id", fmt.Sprintf("%d", msg.ContractId)),
            sdk.NewAttribute("initiator", msg.Creator),
            sdk.NewAttribute("deadline", fmt.Sprintf("%d", deadline)),
        ),
    )
    
    return &types.MsgOpenDisputeResponse{
        DisputeId: disputeId,
    }, nil
}
