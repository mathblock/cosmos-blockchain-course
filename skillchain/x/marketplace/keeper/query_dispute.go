package keeper

import (
	"context"
	"errors"

	"skillchain/x/marketplace/types"

	"cosmossdk.io/collections"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) ListDispute(ctx context.Context, req *types.QueryAllDisputeRequest) (*types.QueryAllDisputeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	disputes, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.Dispute,
		req.Pagination,
		func(_ uint64, value types.Dispute) (types.Dispute, error) {
			return value, nil
		},
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllDisputeResponse{Dispute: disputes, Pagination: pageRes}, nil
}

func (q queryServer) GetDispute(ctx context.Context, req *types.QueryGetDisputeRequest) (*types.QueryGetDisputeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	dispute, err := q.k.Dispute.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrKeyNotFound
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &types.QueryGetDisputeResponse{Dispute: dispute}, nil
}
