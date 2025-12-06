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

func createNProfile(keeper keeper.Keeper, ctx context.Context, n int) []types.Profile {
	items := make([]types.Profile, n)
	for i := range items {
		items[i].Owner = strconv.Itoa(i)
		items[i].Name = strconv.Itoa(i)
		items[i].Bio = strconv.Itoa(i)
		items[i].Skills = []string{`abc` + strconv.Itoa(i), `xyz` + strconv.Itoa(i)}
		items[i].HourlyRate = uint64(i)
		items[i].TotalJobs = uint64(i)
		items[i].TotalEarned = uint64(i)
		items[i].RatingSum = uint64(i)
		items[i].RatingCount = uint64(i)
		_ = keeper.Profile.Set(ctx, items[i].Owner, items[i])
	}
	return items
}

func TestProfileQuerySingle(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNProfile(f.keeper, f.ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetProfileRequest
		response *types.QueryGetProfileResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetProfileRequest{
				Owner: msgs[0].Owner,
			},
			response: &types.QueryGetProfileResponse{Profile: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetProfileRequest{
				Owner: msgs[1].Owner,
			},
			response: &types.QueryGetProfileResponse{Profile: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetProfileRequest{
				Owner: strconv.Itoa(100000),
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
			response, err := qs.GetProfile(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.EqualExportedValues(t, tc.response, response)
			}
		})
	}
}

func TestProfileQueryPaginated(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNProfile(f.keeper, f.ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllProfileRequest {
		return &types.QueryAllProfileRequest{
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
			resp, err := qs.ListProfile(f.ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Profile), step)
			require.Subset(t, msgs, resp.Profile)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListProfile(f.ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Profile), step)
			require.Subset(t, msgs, resp.Profile)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := qs.ListProfile(f.ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.EqualExportedValues(t, msgs, resp.Profile)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := qs.ListProfile(f.ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
