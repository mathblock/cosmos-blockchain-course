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

func (q queryServer) ListApplication(ctx context.Context, req *types.QueryAllApplicationRequest) (*types.QueryAllApplicationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	applications, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.Application,
		req.Pagination,
		func(_ uint64, value types.Application) (types.Application, error) {
			return value, nil
		},
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllApplicationResponse{Application: applications, Pagination: pageRes}, nil
}

func (q queryServer) GetApplication(ctx context.Context, req *types.QueryGetApplicationRequest) (*types.QueryGetApplicationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	application, err := q.k.Application.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrKeyNotFound
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &types.QueryGetApplicationResponse{Application: application}, nil
}
