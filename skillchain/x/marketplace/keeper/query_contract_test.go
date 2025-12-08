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

func createNContract(keeper keeper.Keeper, ctx context.Context, n int) []types.Contract {
	items := make([]types.Contract, n)
	for i := range items {
		iu := uint64(i)
		items[i].Id = iu
		items[i].GigId = uint64(i)
		items[i].ApplicationId = uint64(i)
		items[i].Client = strconv.Itoa(i)
		items[i].Freelancer = strconv.Itoa(i)
		items[i].Price = uint64(i)
		items[i].DeliveryDeadline = int64(i)
		items[i].Status = strconv.Itoa(i)
		items[i].CreatedAt = int64(i)
		items[i].CompletedAt = int64(i)
		_ = keeper.Contract.Set(ctx, iu, items[i])
		_ = keeper.ContractSeq.Set(ctx, iu)
	}
	return items
}

func TestContractQuerySingle(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNContract(f.keeper, f.ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetContractRequest
		response *types.QueryGetContractResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetContractRequest{Id: msgs[0].Id},
			response: &types.QueryGetContractResponse{Contract: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetContractRequest{Id: msgs[1].Id},
			response: &types.QueryGetContractResponse{Contract: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetContractRequest{Id: uint64(len(msgs))},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := qs.GetContract(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.EqualExportedValues(t, tc.response, response)
			}
		})
	}
}

func TestContractQueryPaginated(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNContract(f.keeper, f.ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllContractRequest {
		return &types.QueryAllContractRequest{
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
			resp, err := qs.ListContract(f.ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Contract), step)
			require.Subset(t, msgs, resp.Contract)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListContract(f.ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Contract), step)
			require.Subset(t, msgs, resp.Contract)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := qs.ListContract(f.ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.EqualExportedValues(t, msgs, resp.Contract)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := qs.ListContract(f.ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
