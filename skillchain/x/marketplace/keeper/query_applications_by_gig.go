package keeper

import (
	"context"

	"skillchain/x/marketplace/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) ApplicationsByGig(ctx context.Context, req *types.QueryApplicationsByGigRequest) (*types.QueryApplicationsByGigResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// TODO: Process the query

	return &types.QueryApplicationsByGigResponse{}, nil
}
