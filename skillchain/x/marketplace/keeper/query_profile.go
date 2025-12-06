package keeper

import (
	"context"
	"errors"

	"skillchain/x/marketplace/types"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) ListProfile(ctx context.Context, req *types.QueryAllProfileRequest) (*types.QueryAllProfileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	profiles, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.Profile,
		req.Pagination,
		func(_ string, value types.Profile) (types.Profile, error) {
			return value, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllProfileResponse{Profile: profiles, Pagination: pageRes}, nil
}

func (q queryServer) GetProfile(ctx context.Context, req *types.QueryGetProfileRequest) (*types.QueryGetProfileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, err := q.k.Profile.Get(ctx, req.Owner)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &types.QueryGetProfileResponse{Profile: val}, nil
}
