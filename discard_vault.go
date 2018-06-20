package stl

import "context"

// NewDiscardVault creates an instance of Vault that does nothing.
func NewDiscardVault() Vault {
	return &discardVault{}
}

type discardVault struct {
}

func (v *discardVault) Lock(context.Context, Tx) error {
	return nil
}

func (v *discardVault) Unlock(Tx) {
}
