package keeper

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	cmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/circuit/types"
)

var addresses = []string{
	"cosmos1zglwfu6xjzvzagqcmvzewyzjp9xwqw5qwrr8n9",
	"cosmos1p8s0p6gqc6c9gt77lgr2qqujz49huhu6a80smx",
	"cosmos1qasf9ehx8m7cnat39ndc74rx3fg7z66u8lw0fd",
	"cosmos1uxrdj5zfuudhypsmmjxnj4gpu432ycht06a05a",
}

type fixture struct {
	Ctx        sdk.Context
	Keeper     Keeper
	MockAddr   string
	MockPerms  types.Permissions
	MockMsgURL string
}

func SetupFixture(t *testing.T) *fixture {
	mockStoreKey := storetypes.NewKVStoreKey("test")
	mockAddr := "mock_address"
	keeperX := NewKeeper(mockStoreKey, addresses[0])
	mockMsgURL := "mock_url"
	mockCtx := testutil.DefaultContextWithDB(t, mockStoreKey, storetypes.NewTransientStoreKey("transient_test"))
	ctx := mockCtx.Ctx.WithBlockHeader(cmproto.Header{})
	mockPerms := types.Permissions{
		Level:         3,
		LimitTypeUrls: []string{"test"},
	}

	return &fixture{
		Ctx:        ctx,
		Keeper:     keeperX,
		MockAddr:   mockAddr,
		MockPerms:  mockPerms,
		MockMsgURL: mockMsgURL,
	}
}