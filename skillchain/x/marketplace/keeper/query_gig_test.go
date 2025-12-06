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

func createNGig(keeper keeper.Keeper, ctx context.Context, n int) []types.Gig {
	items := make([]types.Gig, n)
	for i := range items {
		iu := uint64(i)
		items[i].Id = iu
		items[i].Title = strconv.Itoa(i)
		items[i].Description = strconv.Itoa(i)
		items[i].Owner = strconv.Itoa(i)
		items[i].Price = uint64(i)
		items[i].Category = strconv.Itoa(i)
		items[i].DeliveryDays = uint64(i)
		items[i].Status = strconv.Itoa(i)
		items[i].CreatedAt = int64(i)
		_ = keeper.Gig.Set(ctx, iu, items[i])
		_ = keeper.GigSeq.Set(ctx, iu)
	}
	return items
}

func TestGigQuerySingle(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNGig(f.keeper, f.ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetGigRequest
		response *types.QueryGetGigResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetGigRequest{Id: msgs[0].Id},
			response: &types.QueryGetGigResponse{Gig: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetGigRequest{Id: msgs[1].Id},
			response: &types.QueryGetGigResponse{Gig: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetGigRequest{Id: uint64(len(msgs))},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := qs.GetGig(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.EqualExportedValues(t, tc.response, response)
			}
		})
	}
}

func TestGigQueryPaginated(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNGig(f.keeper, f.ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllGigRequest {
		return &types.QueryAllGigRequest{
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
			resp, err := qs.ListGig(f.ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Gig), step)
			require.Subset(t, msgs, resp.Gig)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListGig(f.ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Gig), step)
			require.Subset(t, msgs, resp.Gig)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := qs.ListGig(f.ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.EqualExportedValues(t, msgs, resp.Gig)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := qs.ListGig(f.ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
