package keeper

import (
	"context"
	"fmt"

	"skillchain/x/marketplace/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) AcceptApplication(goCtx context.Context, msg *types.MsgAcceptApplication) (*types.MsgAcceptApplicationResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message
	ctx := sdk.UnwrapSDKContext(goCtx)

	application, err := k.Application.Get(ctx, msg.ApplicationId)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "application with id %d not found: %v", msg.ApplicationId, err)
	}

	if application.Status != "pending" {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"application %d is not pending (status: %s)",
			msg.ApplicationId,
			application.Status,
		)
	}

	gig, err := k.Gig.Get(ctx, application.GigId)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "gig with id %d not found: %v", application.GigId, err)
	}

	if gig.Owner != msg.Creator {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrUnauthorized,
			"only the gig creator (%s) can accept applications",
			gig.Owner,
		)
	}

	if gig.Status != "open" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "gig is no longer open")
	}

	application.Status = "accepted"
	err = k.Application.Set(ctx, application.Id, application)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "failed to update application status: %v", err)
	}

	gig.Status = "in_progress"
	err = k.Gig.Set(ctx, gig.Id, gig)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "failed to update gig status: %v", err)
	}

	err = k.Application.Walk(ctx, nil, func(key uint64, application types.Application) (stop bool, err error) {
		if application.GigId == gig.Id && application.Id != msg.ApplicationId && application.Status == "pending" {
			application.Status = "rejected"
			err = k.Application.Set(ctx, application.Id, application)
		}
		return false, nil
	})
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "failed to reject other applications: %v", err)
	}

	deliveryDeadline := ctx.BlockTime().Unix() + int64(application.ProposedDays*86400)

	contractId, err := k.ContractSeq.Next(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to get next contract id")
	}

	contract := types.Contract{
		Id:               contractId,
		GigId:            gig.Id,
		ApplicationId:    application.Id,
		Client:           gig.Owner,
		Freelancer:       application.Freelancer,
		Price:            application.ProposedPrice,
		DeliveryDeadline: deliveryDeadline,
		Status:           "active",
		CreatedAt:        ctx.BlockTime().Unix(),
		CompletedAt:      0,
	}

	err = k.Contract.Set(ctx, contract.Id, contract)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "failed to create contract: %v", err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"accept_application",
			sdk.NewAttribute("ApplicationId", fmt.Sprintf("%d", application.Id)),
			sdk.NewAttribute("GigId", fmt.Sprintf("%d", gig.Id)),
			sdk.NewAttribute("ContractId", fmt.Sprintf("%d", contract.Id)),
		),
		sdk.NewEvent(
			"contract_created",
			sdk.NewAttribute("contract_id", fmt.Sprintf("%d", contract.Id)),
			sdk.NewAttribute("client", contract.Client),
			sdk.NewAttribute("freelancer", contract.Freelancer),
			sdk.NewAttribute("price", fmt.Sprintf("%d", contract.Price)),
			sdk.NewAttribute("delivery_deadline", fmt.Sprintf("%d", contract.DeliveryDeadline)),
		),
	})

	return &types.MsgAcceptApplicationResponse{
		ContractId: contract.Id,
	}, nil
}
