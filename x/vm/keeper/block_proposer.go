package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetCoinbaseAddress returns the block proposer's validator operator address.
// Returns zero address if any error occurs.
func (k Keeper) GetCoinbaseAddress(ctx sdk.Context, proposerAddress sdk.ConsAddress) (common.Address, error) {
	validator, err := k.stakingKeeper.GetValidatorByConsAddr(ctx, GetProposerAddress(ctx, proposerAddress))
	if err != nil {
		return common.Address{}, nil
	}

	bz, err := sdk.ValAddressFromBech32(validator.GetOperator())
	if err != nil {
		return common.Address{}, errorsmod.Wrapf(
			err,
			"failed to convert validator operator address %s to bytes",
			validator.GetOperator(),
		)
	}
	return common.BytesToAddress(bz), nil
}

// GetProposerAddress returns current block proposer's address when provided proposer address is empty.
func GetProposerAddress(ctx sdk.Context, proposerAddress sdk.ConsAddress) sdk.ConsAddress {
	if len(proposerAddress) == 0 {
		proposerAddress = ctx.BlockHeader().ProposerAddress
	}
	return proposerAddress
}
