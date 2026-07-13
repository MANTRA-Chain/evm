package evm_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/evm/ante/evm"

	"cosmossdk.io/log/v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestEthSigVerificationDecorator_CacheHit(t *testing.T) {
	dec := evm.NewEthSigVerificationDecorator(nil)
	var tx sdk.Tx
	newCtx := func() sdk.Context {
		return sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger()).
			WithIncarnationCache(map[string]any{})
	}
	next := func(c sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) { return c, nil }
	cachedErr := errors.New("cached sig verification failure")

	t.Run("cached error short-circuits", func(t *testing.T) {
		ctx := newCtx()
		ctx.SetIncarnationCache(evm.EthSigVerificationResultCacheKey, cachedErr)
		_, err := dec.AnteHandle(ctx, tx, true, next)
		require.ErrorIs(t, err, cachedErr)
	})

	t.Run("non-error cached value returns explicit error", func(t *testing.T) {
		ctx := newCtx()
		ctx.SetIncarnationCache(evm.EthSigVerificationResultCacheKey, "not-an-error")
		_, err := dec.AnteHandle(ctx, tx, true, next)
		require.ErrorContains(t, err, "unexpected type string")
	})

	t.Run("cached nil success calls next without re-verifying", func(t *testing.T) {
		ctx := newCtx()
		ctx.SetIncarnationCache(evm.EthSigVerificationResultCacheKey, nil)

		nextCalled := false
		_, err := dec.AnteHandle(ctx, tx, true, func(c sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
			nextCalled = true
			return c, nil
		})
		require.NoError(t, err)
		require.True(t, nextCalled, "next handler must run when cache holds a nil success")
	})
}

func TestEthSigVerificationDecorator_RespectsIsSigverifyTx(t *testing.T) {
	// A nil keeper + nil tx would panic if the decorator reached verification, so
	// reaching `next` without error proves the gate short-circuited before any
	// ecrecover or cache access.
	dec := evm.NewEthSigVerificationDecorator(nil)
	var tx sdk.Tx
	next := func(c sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) { return c, nil }

	t.Run("skips verification when IsSigverifyTx is false", func(t *testing.T) {
		ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger()).
			WithIncarnationCache(map[string]any{}).
			WithIsSigverifyTx(false)

		nextCalled := false
		_, err := dec.AnteHandle(ctx, tx, false, func(c sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
			nextCalled = true
			return c, nil
		})
		require.NoError(t, err)
		require.True(t, nextCalled, "next must run when the node opted out of re-verification")
		// nothing was cached, since verification was skipped entirely
		_, ok := ctx.GetIncarnationCache(evm.EthSigVerificationResultCacheKey)
		require.False(t, ok, "no cache entry should be written when verification is skipped")
	})

	t.Run("default (true) does not short-circuit on the gate", func(t *testing.T) {
		// Default context has IsSigverifyTx() == true; a nil cache miss forces the
		// decorator past the gate into verification, which panics on the nil tx --
		// confirming the gate did not swallow the default path.
		ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger()).
			WithIncarnationCache(map[string]any{})
		require.True(t, ctx.IsSigverifyTx())
		require.Panics(t, func() {
			_, _ = dec.AnteHandle(ctx, tx, false, next)
		})
	})
}
