package keeper

import (
	"context"

	"skillchain/x/marketplace/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) EscrowBalance(goCtx context.Context, req *types.QueryEscrowBalanceRequest) (*types.QueryEscrowBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// TODO: Process the query
	ctx := sdk.UnwrapSDKContext(goCtx)

	addr := q.k.accountKeeper.GetModuleAddress(types.ModuleName)
	balance := q.k.bankKeeper.GetBalance(ctx, addr, "skill")

	return &types.QueryEscrowBalanceResponse{Balance: &balance}, nil
}
