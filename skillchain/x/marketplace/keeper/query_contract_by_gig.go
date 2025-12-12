package keeper

import (
	"context"

	"skillchain/x/marketplace/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) ContractByGig(goCtx context.Context, req *types.QueryContractByGigRequest) (*types.QueryContractByGigResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// TODO: Process the query
	ctx := sdk.UnwrapSDKContext(goCtx)

	contract, err := q.k.Contract.Get(ctx, req.GigId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "contract not found")
	}

	return &types.QueryContractByGigResponse{Contract: &contract}, nil
}
