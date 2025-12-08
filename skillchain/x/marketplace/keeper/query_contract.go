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

func (q queryServer) ListContract(ctx context.Context, req *types.QueryAllContractRequest) (*types.QueryAllContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	contracts, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.Contract,
		req.Pagination,
		func(_ uint64, value types.Contract) (types.Contract, error) {
			return value, nil
		},
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllContractResponse{Contract: contracts, Pagination: pageRes}, nil
}

func (q queryServer) GetContract(ctx context.Context, req *types.QueryGetContractRequest) (*types.QueryGetContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	contract, err := q.k.Contract.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrKeyNotFound
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &types.QueryGetContractResponse{Contract: contract}, nil
}
