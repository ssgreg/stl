package stl

import (
	"context"
	"sync"
)

// A Locker represents an object that can be locked and unlocked.
type Locker interface {
	sync.Locker

	LockWithContext(context.Context) error
}

// Tx holds the names of shared and exclusive resources to be locked
// atomically.
type Tx interface {
	ListShared() []string
	ListExclusive() []string
}

// Builder allows to build a Locker or a Tx.
type Builder interface {
	Shared(string) Builder
	Exclusive(string) Builder
	ToLocker(Vault) Locker
	ToTx() Tx
}

// Vault defines a resource vault that can lock/unlock resources according to
// the specified transaction.
type Vault interface {
	Lock(context.Context, Tx) error
	Unlock(Tx)
}
