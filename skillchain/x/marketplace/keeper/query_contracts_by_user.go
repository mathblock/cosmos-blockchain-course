package keeper

import (
	"context"

	"skillchain/x/marketplace/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) ContractsByUser(goCtx context.Context, req *types.QueryContractsByUserRequest) (*types.QueryContractsByUserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// TODO: Process the query
	ctx := sdk.UnwrapSDKContext(goCtx)

	var contracts []types.Contract
	err := q.k.Contract.Walk(ctx, nil, func(key uint64, contract types.Contract) (stop bool, err error) {
		if contract.Client == req.User || contract.Freelancer == req.User {
			contracts = append(contracts, contract)
		}
		return false, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to retrieve contracts")
	}

	return &types.QueryContractsByUserResponse{Contracts: contracts}, nil
}
