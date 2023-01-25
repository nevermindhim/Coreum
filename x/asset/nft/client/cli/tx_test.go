package cli_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/asset/nft/client/cli"
	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

func TestCmdTxIssueClass(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	args := []string{
		symbol,
		"class name",
		"class description",
		"https://my-class-meta.invalid/1",
		"content-hash",
		fmt.Sprintf("--features=%s", types.ClassFeature_burning.String()), //nolint:nosnakecase // generated variable
	}
	args = append(args, txValidator1Args(testNetwork)...)
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueClass(), args)
	requireT.NoError(err)

	var res sdk.TxResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &res))
	requireT.NotEmpty(res.TxHash)
	requireT.Equal(uint32(0), res.Code, "can't submit IssueClass tx", res)
}

func TestCmdFreeze(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	// create class
	classID := issueClass(
		requireT,
		ctx,
		symbol,
		"class name",
		"class description",
		"https://my-class-meta.invalid/1",
		"",
		testNetwork,
		types.ClassFeature_freezing, //nolint:nosnakecase // generated variable
	)
	// mint nft
	nftID := "nft-1"
	mint(
		requireT,
		ctx,
		classID,
		nftID,
		"https://my-nft-meta.invalid/1",
		"9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9",
		testNetwork,
	)

	// freeze
	args := []string{classID, nftID}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxFreeze(), args)
	requireT.NoError(err)

	// query frozen
	var frozenResp types.QueryFrozenResponse
	args = []string{classID, nftID}
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryFrozen(), args)
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &frozenResp))
	requireT.True(frozenResp.Frozen)

	// unfreeze
	args = []string{classID, nftID}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxUnfreeze(), args)
	requireT.NoError(err)

	// query frozen
	args = []string{classID, nftID}
	buf, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryFrozen(), args)
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &frozenResp))
	requireT.False(frozenResp.Frozen)
}

func txValidator1Args(testNetwork *network.Network) []string {
	return []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, testNetwork.Validators[0].Address.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(testNetwork.Config.BondDenom, sdk.NewInt(1000000))).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
}