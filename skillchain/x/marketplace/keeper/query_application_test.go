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

func createNApplication(keeper keeper.Keeper, ctx context.Context, n int) []types.Application {
	items := make([]types.Application, n)
	for i := range items {
		iu := uint64(i)
		items[i].Id = iu
		items[i].GigId = uint64(i)
		items[i].Freelancer = strconv.Itoa(i)
		items[i].CoverLetter = strconv.Itoa(i)
		items[i].ProposedPrice = uint64(i)
		items[i].ProposedDays = uint64(i)
		items[i].Status = strconv.Itoa(i)
		items[i].CreatedAt = int64(i)
		_ = keeper.Application.Set(ctx, iu, items[i])
		_ = keeper.ApplicationSeq.Set(ctx, iu)
	}
	return items
}

func TestApplicationQuerySingle(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNApplication(f.keeper, f.ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetApplicationRequest
		response *types.QueryGetApplicationResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetApplicationRequest{Id: msgs[0].Id},
			response: &types.QueryGetApplicationResponse{Application: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetApplicationRequest{Id: msgs[1].Id},
			response: &types.QueryGetApplicationResponse{Application: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetApplicationRequest{Id: uint64(len(msgs))},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := qs.GetApplication(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.EqualExportedValues(t, tc.response, response)
			}
		})
	}
}

func TestApplicationQueryPaginated(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNApplication(f.keeper, f.ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllApplicationRequest {
		return &types.QueryAllApplicationRequest{
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
			resp, err := qs.ListApplication(f.ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Application), step)
			require.Subset(t, msgs, resp.Application)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListApplication(f.ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Application), step)
			require.Subset(t, msgs, resp.Application)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := qs.ListApplication(f.ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.EqualExportedValues(t, msgs, resp.Application)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := qs.ListApplication(f.ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
