package keeper

import (
	"context"
	"errors"
	"fmt"

	"skillchain/x/marketplace/types"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CreateApplication(ctx context.Context, msg *types.MsgCreateApplication) (*types.MsgCreateApplicationResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid address: %s", err))
	}

	nextId, err := k.ApplicationSeq.Next(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to get next id")
	}

	var application = types.Application{
		Id:            nextId,
		Creator:       msg.Creator,
		GigId:         msg.GigId,
		Freelancer:    msg.Freelancer,
		CoverLetter:   msg.CoverLetter,
		ProposedPrice: msg.ProposedPrice,
		ProposedDays:  msg.ProposedDays,
		Status:        msg.Status,
		CreatedAt:     msg.CreatedAt,
	}

	if err = k.Application.Set(
		ctx,
		nextId,
		application,
	); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to set application")
	}

	return &types.MsgCreateApplicationResponse{
		Id: nextId,
	}, nil
}

func (k msgServer) UpdateApplication(ctx context.Context, msg *types.MsgUpdateApplication) (*types.MsgUpdateApplicationResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid address: %s", err))
	}

	var application = types.Application{
		Creator:       msg.Creator,
		Id:            msg.Id,
		GigId:         msg.GigId,
		Freelancer:    msg.Freelancer,
		CoverLetter:   msg.CoverLetter,
		ProposedPrice: msg.ProposedPrice,
		ProposedDays:  msg.ProposedDays,
		Status:        msg.Status,
		CreatedAt:     msg.CreatedAt,
	}

	// Checks that the element exists
	val, err := k.Application.Get(ctx, msg.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
		}

		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to get application")
	}

	// Checks if the msg creator is the same as the current owner
	if msg.Creator != val.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	if err := k.Application.Set(ctx, msg.Id, application); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to update application")
	}

	return &types.MsgUpdateApplicationResponse{}, nil
}

func (k msgServer) DeleteApplication(ctx context.Context, msg *types.MsgDeleteApplication) (*types.MsgDeleteApplicationResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid address: %s", err))
	}

	// Checks that the element exists
	val, err := k.Application.Get(ctx, msg.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
		}

		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to get application")
	}

	// Checks if the msg creator is the same as the current owner
	if msg.Creator != val.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	if err := k.Application.Remove(ctx, msg.Id); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to delete application")
	}

	return &types.MsgDeleteApplicationResponse{}, nil
}
