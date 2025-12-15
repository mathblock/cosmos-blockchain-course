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

func (q queryServer) ListDisputeVote(ctx context.Context, req *types.QueryAllDisputeVoteRequest) (*types.QueryAllDisputeVoteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	disputeVotes, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.DisputeVote,
		req.Pagination,
		func(_ string, value types.DisputeVote) (types.DisputeVote, error) {
			return value, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllDisputeVoteResponse{DisputeVote: disputeVotes, Pagination: pageRes}, nil
}

func (q queryServer) GetDisputeVote(ctx context.Context, req *types.QueryGetDisputeVoteRequest) (*types.QueryGetDisputeVoteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, err := q.k.DisputeVote.Get(ctx, req.Arbiter)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &types.QueryGetDisputeVoteResponse{DisputeVote: val}, nil
}
