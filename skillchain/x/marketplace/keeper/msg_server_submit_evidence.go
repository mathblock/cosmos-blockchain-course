package keeper

import (
	"context"
	"fmt"

	"skillchain/x/marketplace/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) SubmitEvidence(goCtx context.Context, msg *types.MsgSubmitEvidence) (*types.MsgSubmitEvidenceResponse, error) {
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
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "dispute is not open for evidence")
    }
    
    if ctx.BlockTime().Unix() > dispute.Deadline {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "dispute deadline has passed")
    }
    
    contract, err := k.Contract.Get(ctx, dispute.ContractId)
    if err != nil {
        return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "contract %d not found", dispute.ContractId)
    }
    
    isClient := contract.Client == msg.Creator
    isFreelancer := contract.Freelancer == msg.Creator
    
    if !isClient && !isFreelancer {
        return nil, errorsmod.Wrap(types.ErrUnauthorized, "only parties can submit evidence")
    }
    
    if isClient {
        if dispute.ClientEvidence != "" {
            return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "client evidence already submitted")
        }
        dispute.ClientEvidence = msg.Evidence
    } else {
        if dispute.FreelancerEvidence != "" {
            return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "freelancer evidence already submitted")
        }
        dispute.FreelancerEvidence = msg.Evidence
    }
    
    if dispute.ClientEvidence != "" && dispute.FreelancerEvidence != "" {
        dispute.Status = "voting"
    }
    
    err = k.Dispute.Set(ctx, dispute.Id, dispute)
    if err != nil {
        return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "failed to update dispute: %v", err)
    }
    
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "evidence_submitted",
            sdk.NewAttribute("dispute_id", fmt.Sprintf("%d", msg.DisputeId)),
            sdk.NewAttribute("submitter", msg.Creator),
        ),
    )

	return &types.MsgSubmitEvidenceResponse{}, nil
}
