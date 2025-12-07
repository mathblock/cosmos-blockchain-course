package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"skillchain/x/marketplace/keeper"
	"skillchain/x/marketplace/types"
)

func TestMsgCreateGig(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	// Create first gig
	msg1 := &types.MsgCreateGig{
		Creator:      "skill13axcg4tlh0e6efuytpd52n5w5f2sv3xv6esu4x",
		Title:        "First Gig Title",
		Description:  "This is the first gig description which is long enough.",
		Price:        100,
		Category:     "development",
		DeliveryDays: 10,
	}

	res1, err := ms.CreateGig(f.ctx, msg1)
	require.NoError(t, err)
	require.Equal(t, uint64(0), res1.Id)

	// Create second gig
	msg2 := &types.MsgCreateGig{
		Creator:      "skill13axcg4tlh0e6efuytpd52n5w5f2sv3xv6esu4x",
		Title:        "Second Gig Title",
		Description:  "This is the second gig description which is long enough.",
		Price:        200,
		Category:     "design",
		DeliveryDays: 5,
	}

	res2, err := ms.CreateGig(f.ctx, msg2)
	require.NoError(t, err)
	require.Equal(t, uint64(1), res2.Id)

	// Verify gigs are stored correctly
	gig1, err := f.keeper.Gig.Get(f.ctx, res1.Id)
	require.NoError(t, err)
	require.Equal(t, msg1.Title, gig1.Title)

	gig2, err := f.keeper.Gig.Get(f.ctx, res2.Id)
	require.NoError(t, err)
	require.Equal(t, msg2.Title, gig2.Title)
}
