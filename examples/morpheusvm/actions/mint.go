package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/math"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	mconsts "github.com/ava-labs/hypersdk/examples/morpheusvm/consts"
	"github.com/ava-labs/hypersdk/examples/morpheusvm/storage"
	"github.com/ava-labs/hypersdk/state"
)

var _ chain.Action = (*Mint)(nil)

type Mint struct {
	To    codec.Address `json:"to"`
	Value uint64        `json:"value"`
}

func (*Mint) GetTypeID() uint8 {
	return mconsts.MintID
}

func (m *Mint) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	return state.Keys{
		string(storage.TotalSupplyKey()): state.Read | state.Write,
		string(storage.BalanceKey(m.To)): state.All,
	}
}

func (*Mint) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.TotalSupplyChunks, storage.BalanceChunks}
}

func (m Mint) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (bool, uint64, []byte, error) {

	// Check if the amount is valid
	if m.Value == 0 {
		return false, 1, []byte("Value cannot be 0"), nil
	}

	//Update the total supply
	totalSupply, err := storage.GetTotalSupply(ctx, mu)
	if err != nil {
		return false, 1, []byte(err.Error()), nil
	}

	newTotalSupply, err := math.Add64(totalSupply, m.Value)
	if err != nil {
		return false, 1, []byte(err.Error()), nil
	}

	if err := storage.SetTotalSupply(ctx, mu, newTotalSupply); err != nil {
		return false, 1, []byte(err.Error()), nil
	}

	// Update the balance of the recipient
	if err := storage.AddBalance(ctx, mu, m.To, m.Value, true); err != nil {
		return false, 1, []byte(err.Error()), nil
	}

	return true, 1, nil, nil

}

func (*Mint) MaxComputeUnits(chain.Rules) uint64 {
	return MintComputeUnits
}

func (*Mint) Size() int {
	return codec.AddressLen + consts.Uint64Len
}

func (m *Mint) Marshal(p *codec.Packer) {
	p.PackAddress(m.To)
	p.PackUint64(m.Value)
}

func UnmarshalMint(p *codec.Packer) (chain.Action, error) {
	var mint Mint
	p.UnpackAddress(&mint.To) // we do not verify the typeID is valid
	mint.Value = p.UnpackUint64(true)
	if err := p.Err(); err != nil {
		return nil, err
	}
	return &mint, nil
}

func (*Mint) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
