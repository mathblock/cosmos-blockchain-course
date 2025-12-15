package keeper_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"skillchain/x/marketplace/keeper"
	"skillchain/x/marketplace/types"
)

func createNDisputeVote(keeper keeper.Keeper, ctx context.Context, n int) []types.DisputeVote {
	items := make([]types.DisputeVote, n)
	for i := range items {
		items[i].Arbiter = strconv.Itoa(i)
		items[i].DisputeId = uint64(i)
		items[i].Vote = strconv.Itoa(i)
		items[i].VotedAt = int64(i)
		_ = keeper.DisputeVote.Set(ctx, items[i].Arbiter, items[i])
	}
	return items
}

func TestDisputeVoteQuerySingle(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNDisputeVote(f.keeper, f.ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetDisputeVoteRequest
		response *types.QueryGetDisputeVoteResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetDisputeVoteRequest{
				Arbiter: msgs[0].Arbiter,
			},
			response: &types.QueryGetDisputeVoteResponse{DisputeVote: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetDisputeVoteRequest{
				Arbiter: msgs[1].Arbiter,
			},
			response: &types.QueryGetDisputeVoteResponse{DisputeVote: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetDisputeVoteRequest{
				Arbiter: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := qs.GetDisputeVote(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.EqualExportedValues(t, tc.response, response)
			}
		})
	}
}

func TestDisputeVoteQueryPaginated(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNDisputeVote(f.keeper, f.ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllDisputeVoteRequest {
		return &types.QueryAllDisputeVoteRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListDisputeVote(f.ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.DisputeVote), step)
			require.Subset(t, msgs, resp.DisputeVote)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListDisputeVote(f.ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.DisputeVote), step)
			require.Subset(t, msgs, resp.DisputeVote)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := qs.ListDisputeVote(f.ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.EqualExportedValues(t, msgs, resp.DisputeVote)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := qs.ListDisputeVote(f.ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
