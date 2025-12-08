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

func (k msgServer) CreateContract(ctx context.Context, msg *types.MsgCreateContract) (*types.MsgCreateContractResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid address: %s", err))
	}

	nextId, err := k.ContractSeq.Next(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to get next id")
	}

	var contract = types.Contract{
		Id:               nextId,
		Creator:          msg.Creator,
		GigId:            msg.GigId,
		ApplicationId:    msg.ApplicationId,
		Client:           msg.Client,
		Freelancer:       msg.Freelancer,
		Price:            msg.Price,
		DeliveryDeadline: msg.DeliveryDeadline,
		Status:           msg.Status,
		CreatedAt:        msg.CreatedAt,
		CompletedAt:      msg.CompletedAt,
	}

	if err = k.Contract.Set(
		ctx,
		nextId,
		contract,
	); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to set contract")
	}

	return &types.MsgCreateContractResponse{
		Id: nextId,
	}, nil
}

func (k msgServer) UpdateContract(ctx context.Context, msg *types.MsgUpdateContract) (*types.MsgUpdateContractResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid address: %s", err))
	}

	var contract = types.Contract{
		Creator:          msg.Creator,
		Id:               msg.Id,
		GigId:            msg.GigId,
		ApplicationId:    msg.ApplicationId,
		Client:           msg.Client,
		Freelancer:       msg.Freelancer,
		Price:            msg.Price,
		DeliveryDeadline: msg.DeliveryDeadline,
		Status:           msg.Status,
		CreatedAt:        msg.CreatedAt,
		CompletedAt:      msg.CompletedAt,
	}

	// Checks that the element exists
	val, err := k.Contract.Get(ctx, msg.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
		}

		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to get contract")
	}

	// Checks if the msg creator is the same as the current owner
	if msg.Creator != val.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	if err := k.Contract.Set(ctx, msg.Id, contract); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to update contract")
	}

	return &types.MsgUpdateContractResponse{}, nil
}

func (k msgServer) DeleteContract(ctx context.Context, msg *types.MsgDeleteContract) (*types.MsgDeleteContractResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid address: %s", err))
	}

	// Checks that the element exists
	val, err := k.Contract.Get(ctx, msg.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
		}

		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to get contract")
	}

	// Checks if the msg creator is the same as the current owner
	if msg.Creator != val.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	if err := k.Contract.Remove(ctx, msg.Id); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to delete contract")
	}

	return &types.MsgDeleteContractResponse{}, nil
}
