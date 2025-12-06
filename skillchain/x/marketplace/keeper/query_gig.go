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

func (q queryServer) ListGig(ctx context.Context, req *types.QueryAllGigRequest) (*types.QueryAllGigResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	gigs, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.Gig,
		req.Pagination,
		func(_ uint64, value types.Gig) (types.Gig, error) {
			return value, nil
		},
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllGigResponse{Gig: gigs, Pagination: pageRes}, nil
}

func (q queryServer) GetGig(ctx context.Context, req *types.QueryGetGigRequest) (*types.QueryGetGigResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	gig, err := q.k.Gig.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrKeyNotFound
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &types.QueryGetGigResponse{Gig: gig}, nil
}
