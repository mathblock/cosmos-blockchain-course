package keeper

import (
	"context"

	"skillchain/x/marketplace/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) ApplicationsByGig(goCtx context.Context, req *types.QueryApplicationsByGigRequest) (*types.QueryApplicationsByGigResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var applications []types.Application

	// Use Walk with nil ranger to iterate all items
	err := q.k.Application.Walk(ctx, nil, func(key uint64, application types.Application) (stop bool, err error) {
		// Filter by GigId during iteration (more efficient)
		if application.GigId == req.GigId {
			applications = append(applications, application)
		}
		return false, nil // Continue iterating
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list applications")
	}

	// Return the filtered applications
	return &types.QueryApplicationsByGigResponse{Applications: applications}, nil
}
