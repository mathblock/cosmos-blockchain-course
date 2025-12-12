package keeper

import (
	"context"

	"skillchain/x/marketplace/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) ApplicationsByFreelancer(goCtx context.Context, req *types.QueryApplicationsByFreelancerRequest) (*types.QueryApplicationsByFreelancerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// TODO: Process the query
	ctx := sdk.UnwrapSDKContext(goCtx)

	var applications []types.Application
	err := q.k.Application.Walk(ctx, nil, func(key uint64, application types.Application) (stop bool, err error) {
		if application.Freelancer == req.Freelancer {
			applications = append(applications, application)
		}
		return false, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to retrieve applications")
	}

	return &types.QueryApplicationsByFreelancerResponse{Applications: applications}, nil
}
