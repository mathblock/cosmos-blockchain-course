package keeper_test

import (
	"context"
	"strconv"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"skillchain/x/marketplace/keeper"
	"skillchain/x/marketplace/types"
)

func createNDispute(keeper keeper.Keeper, ctx context.Context, n int) []types.Dispute {
	items := make([]types.Dispute, n)
	for i := range items {
		iu := uint64(i)
		items[i].Id = iu
		items[i].ContractId = uint64(i)
		items[i].Initiator = strconv.Itoa(i)
		items[i].Reason = strconv.Itoa(i)
		items[i].ClientEvidence = strconv.Itoa(i)
		items[i].FreelancerEvidence = strconv.Itoa(i)
		items[i].Status = strconv.Itoa(i)
		items[i].VotesClient = uint64(i)
		items[i].VotesFreelancer = uint64(i)
		items[i].Resolution = strconv.Itoa(i)
		items[i].CreatedAt = int64(i)
		items[i].Deadline = int64(i)
		_ = keeper.Dispute.Set(ctx, iu, items[i])
		_ = keeper.DisputeSeq.Set(ctx, iu)
	}
	return items
}

func TestDisputeQuerySingle(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNDispute(f.keeper, f.ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetDisputeRequest
		response *types.QueryGetDisputeResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetDisputeRequest{Id: msgs[0].Id},
			response: &types.QueryGetDisputeResponse{Dispute: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetDisputeRequest{Id: msgs[1].Id},
			response: &types.QueryGetDisputeResponse{Dispute: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetDisputeRequest{Id: uint64(len(msgs))},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := qs.GetDispute(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.EqualExportedValues(t, tc.response, response)
			}
		})
	}
}

func TestDisputeQueryPaginated(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNDispute(f.keeper, f.ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllDisputeRequest {
		return &types.QueryAllDisputeRequest{
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
			resp, err := qs.ListDispute(f.ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Dispute), step)
			require.Subset(t, msgs, resp.Dispute)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListDispute(f.ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Dispute), step)
			require.Subset(t, msgs, resp.Dispute)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := qs.ListDispute(f.ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.EqualExportedValues(t, msgs, resp.Dispute)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := qs.ListDispute(f.ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
